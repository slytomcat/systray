extern void systray_ready();
extern void systray_on_exit();
extern void systray_menu_item_selected(int menu_id);
extern void submenu_item_selected(int menuId, int subId);

int nativeLoop(void);

void setIcon(char* icon_file_name);
void setTitle(char* title);
void setTooltip(char* tooltip);
void add_or_update_menu_item(int menuId, char* title, char* tooltip, short disabled, short checked);
void add_separator(int menuId);
void hide_menu_item(int menuId);
void show_menu_item(int menuId);
void quit();

void add_submenu_item(int menuId, int subId, char* title, short disabled);
void remove_submenu(int menuId);
