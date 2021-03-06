// +build !windows

package systray

/*
#cgo linux pkg-config: gtk+-3.0 appindicator3-0.1
#cgo darwin CFLAGS: -DDARWIN -x objective-c -fobjc-arc
#cgo darwin LDFLAGS: -framework Cocoa

#include "systray.h"
*/
import "C"
//import "unsafe"

func nativeLoop() {
	C.nativeLoop()
}

func quit() {
	C.quit()
}

// SetIcon sets the systray icon.
// iconFile should be path to .ico for windows and to .ico/.jpg/.png
// for other platforms.
func SetIcon(iconFile string) {
	C.setIcon(C.CString(iconFile))
}

// SetTitle sets the systray title, available on Linux and on Mac.
func SetTitle(title string) {
	C.setTitle(C.CString(title))
}

// SetTooltip sets the systray tooltip to display on mouse hover of the tray icon,
// only available on Mac and Windows.
func SetTooltip(tooltip string) {
	C.setTooltip(C.CString(tooltip))
}

func addOrUpdateMenuItem(item *MenuItem) {
	var disabled C.short
	if item.disabled {
		disabled = 1
	}
	var checked C.short
	if item.checked {
		checked = 1
	}
	C.add_or_update_menu_item(
		C.int(item.id),
		C.CString(item.title),
		C.CString(item.tooltip),
		disabled,
		checked,
	)
}

func addSeparator(id int32) {
	C.add_separator(C.int(id))
}

func addSubmenuItem(menuId int32, subId int, title string, disabled bool) {
	var cDisabled C.short
	if disabled {
		cDisabled = 1
	}
	C.add_submenu_item(
		C.int(menuId),
		C.int(subId),
		C.CString(title),
		cDisabled,
	)
}

func removeSubmenu(menuId int32){
	C.remove_submenu(C.int(menuId))
}

//export submenu_item_selected
func submenu_item_selected(menuId C.int, subId C.int) {
	submenuItemSelected(int32(menuId), int32(subId))
}


func hideMenuItem(item *MenuItem) {
	C.hide_menu_item(
		C.int(item.id),
	)
}

func showMenuItem(item *MenuItem) {
	C.show_menu_item(
		C.int(item.id),
	)
}

//export systray_ready
func systray_ready() {
	systrayReady()
}

//export systray_on_exit
func systray_on_exit() {
	systrayExit()
}

//export systray_menu_item_selected
func systray_menu_item_selected(cID C.int) {
	systrayMenuItemSelected(int32(cID))
}
