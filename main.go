package main

/*
   #cgo CFLAGS: -I./libvterm/
   #cgo LDFLAGS: ./libvterm/libvterm.a
   #include "vterm.h"
   typedef int (*callback_fcn1)(VTermRect, void*);
   int screenDamage(VTermRect r, void *u); // Forward declaration.
*/
import "C"

import "fmt"
import "reflect"
import "unsafe"

//export screenDamage
func screenDamage(rect C.VTermRect, user unsafe.Pointer) C.int {
	fmt.Printf("damage %d..%d,%d..%d\n", rect.start_row, rect.end_row, rect.start_col, rect.end_col)
	return 0
}

func main() {
	fmt.Printf("CREATE\n")

	var cbs C.VTermScreenCallbacks

	cbs.damage = (C.callback_fcn1)(unsafe.Pointer(C.screenDamage))

	vt := C.vterm_new(25, 80)
	screen := C.vterm_obtain_screen(vt)
	C.vterm_screen_reset(screen, 1)
	C.vterm_screen_enable_altscreen(screen, 1)
	C.vterm_screen_set_callbacks(screen, &cbs, nil)

	C.vterm_input_write(vt, C.CString("lol"), 3)

	fmt.Printf("> %s %s\n", reflect.TypeOf(vt), vt)
	fmt.Printf("> %s %s\n", reflect.TypeOf(screen), screen)

	fmt.Printf("DONE\n")
}
