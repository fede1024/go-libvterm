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
	"bytes"
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

//func cellToStringI

func printRect(screen *C.VTermScreen, rect TermRect) {
	cell := C.VTermScreenCell{}
	pos := C.VTermPos{}
	fmt.Printf("+")
	for c := rect.start_col; c < rect.end_col; c++ {
		fmt.Printf("-")
	}
	fmt.Println("+")

	for r := rect.start_row; r < rect.end_row; r++ {
		fmt.Printf("|")
		for c := rect.start_col; c < rect.end_col; {
			pos.col = C.int(c)
			pos.row = C.int(r)
			C.vterm_screen_get_cell(screen, pos, &cell)
			//str := make([]int, C.VTERM_MAX_CHARS_PER_CELL)
			var buf bytes.Buffer
			for i := 0; i < C.VTERM_MAX_CHARS_PER_CELL && cell.chars[i] != 0; i++ {
				buf.WriteRune((rune)(cell.chars[i]))
			}
			c += int(cell.width)
			if buf.Len() == 0 {
				fmt.Print(" ")
			} else {
				fmt.Print(buf.String())
			}
		}
		fmt.Println("|")
	}

	fmt.Printf("+")
	for c := rect.start_col; c < rect.end_col; c++ {
		fmt.Printf("-")
	}
	fmt.Println("+")
}

func main() {
	fmt.Printf("CREATE\n")

	term := Term{make(chan TermRect), make(chan TermPos), make(chan string),
		make(chan int)}

	var cbs C.VTermScreenCallbacks
	cbs.damage = (C.screenDamage_type)(unsafe.Pointer(C.screenDamage))
	cbs.movecursor = (C.moveCursor_type)(unsafe.Pointer(C.moveCursor))

	vt := C.vterm_new(25, 80)
	C.vterm_set_utf8(vt, 1)
	screen := C.vterm_obtain_screen(vt)
	C.vterm_screen_reset(screen, 1)
	C.vterm_screen_enable_altscreen(screen, 1)
	C.vterm_screen_set_callbacks(screen, &cbs, unsafe.Pointer(&term))

	go func() {
		for {
			select {
			case rect := <-term.damage:
				fmt.Printf("damage: %d\n", rect)
				//printRect(screen, rect)
			case pos := <-term.cursorMove:
				fmt.Printf("move cursor: %d\n", pos)
			case <-term.stop:
				fmt.Printf("stop\n")
				return
			}
		}
	}()

	go func() {
		for str := range term.inputStr {
			C.vterm_input_write(vt, C.CString(str), (C.size_t)(len(str)))
		}
	}()

	//term.inputStr <- "\u2603"
	fmt.Println("\xC3\x81\xC3\xA9")
	fmt.Println("\u2603")
	//term.inputStr <- "\xC3\x81\xC3\xA9"
	term.inputStr <- "\xC3\x81\xC3\xA9"
	//term.inputStr <- "\x1b[H"
	term.inputStr <- "p\xCC\x81"
	term.inputStr <- "a\xCC\x8A"
	term.inputStr <- "zero\xEF\xBC\x90lol"
	term.inputStr <- "\n\rzero\xEF\xBC\x90lol"

	C.vterm_keyboard_unichar(vt, '\u2603', 0)
	bufLen := C.vterm_output_get_buffer_current(vt)
	outbuff := make([]byte, bufLen)
	C.vterm_output_read(vt, (*C.char)(unsafe.Pointer(&outbuff[0])), bufLen)

	term.inputStr <- string(outbuff)

	time.Sleep(100 * time.Millisecond)

	fmt.Println("")
	rect2 := TermRect{0, 0, 5, 20}
	printRect(screen, rect2)

	fmt.Printf("DONE\n")
}
