package main

/*
   #cgo CFLAGS: -I./libvterm/
   #cgo LDFLAGS: ./libvterm/libvterm.a
   #include "vterm.h"
   typedef int (*screenDamage_type)(VTermRect, void*);
   int screenDamage(VTermRect, void *);
   typedef int (*moveCursor_type)(VTermRect, void*);
   int moveCursor(VTermPos, VTermPos, int, void *);
*/
import "C"

import (
	"fmt"
	"time"
	"unsafe"
)

type TermPos struct {
	row int
	col int
}

type TermRect struct {
	start_row int
	start_col int
	end_row   int
	end_col   int
}

type Term struct {
	damage     chan (TermRect)
	cursorMove chan (TermPos)
	inputStr   chan (string)
	stop       chan (int)
}

//var term Term

//export screenDamage
func screenDamage(rect C.VTermRect, user unsafe.Pointer) C.int {
	term := (*Term)(user)
	term.damage <- TermRect{(int)(rect.start_row), (int)(rect.start_col),
		(int)(rect.end_row), (int)(rect.end_col)}

	return 0
}

//export moveCursor
func moveCursor(pos C.VTermPos, oldPos C.VTermPos, visible C.int, user unsafe.Pointer) C.int {
	term := (*Term)(user)
	term.cursorMove <- TermPos{(int)(pos.row), (int)(pos.col)}
	return 0
}

func main() {
	fmt.Printf("CREATE\n")

	term := Term{make(chan TermRect), make(chan TermPos), make(chan string),
		make(chan int)}

	var cbs C.VTermScreenCallbacks
	cbs.damage = (C.screenDamage_type)(unsafe.Pointer(C.screenDamage))
	cbs.movecursor = (C.moveCursor_type)(unsafe.Pointer(C.moveCursor))

	go func() {
		for {
			select {
			case rect := <-term.damage:
				fmt.Printf("damage: %d\n", rect)
			case pos := <-term.cursorMove:
				fmt.Printf("move cursor: %d\n", pos)
			case <-term.stop:
				fmt.Printf("stop\n")
				return
			}
		}
	}()

	vt := C.vterm_new(25, 80)
	screen := C.vterm_obtain_screen(vt)
	C.vterm_screen_reset(screen, 1)
	C.vterm_screen_enable_altscreen(screen, 1)
	C.vterm_screen_set_callbacks(screen, &cbs, unsafe.Pointer(&term))

	go func() {
		for str := range term.inputStr {
			C.vterm_input_write(vt, C.CString(str), (C.size_t)(len(str)))
		}
	}()

	term.inputStr <- "lol"
	term.inputStr <- "test"

	time.Sleep(100 * time.Millisecond)

	fmt.Printf("DONE\n")
}
