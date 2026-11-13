// markdown-viewer-gtk
//
// modified from one-window-browser (added librenderinghtml)
// See: https://github.com/Alkemic/one-window-browser/blob/master/LICENSE

#include <gtk/gtk.h>
#include <librenderinghtml.h> // -I../../librenderinghtml/bin
#include <webkit/webkit.h>

char default_title[] = "testing 1 2 3";
char default_content[] = R"md(# It Works!
```c++
int main (int argc, char *argv[]){
	char *htmlout = render_markdown(title, content);
	fprintf(stdout, "%s\n", htmlout);
	return 0;
}

```

<button onclick="window.webkit.messageHandlers.nativeApp.postMessage('hello_from_html')">
  Click Me
</button>
<button onclick="window.webkit.messageHandlers.nativeApp.postMessage('exit')">
  Exit Button
</button>
)md";

static void on_message(WebKitUserContentManager *,
                       JSCValue *value,
                       gpointer user_data) {
  auto *window = GTK_WINDOW(user_data);
  gchar *message = jsc_value_to_string(value);

  if (g_strcmp0(message, "hello_from_html") == 0) {
    fprintf(stderr, "Clicked button\n");
  } else if (g_strcmp0(message, "exit") == 0) {
    fprintf(stderr, "Clicked exit button\n");
    fflush(stderr);
    g_application_quit(G_APPLICATION(gtk_window_get_application(window)));

  } else {
    fprintf(stderr, "Clicked other: %s\n", message);
  }

  g_free(message);
}
static void activate(GtkApplication *app, gpointer user_data) {
  GtkWidget *window = gtk_application_window_new(app);
  gtk_window_set_title(GTK_WINDOW(window), "markdown-viewer-gtk");
  gtk_window_set_default_size(GTK_WINDOW(window), 1024, 768);
  WebKitWebView *web = WEBKIT_WEB_VIEW(webkit_web_view_new());
  WebKitSettings *settings = webkit_web_view_get_settings(web);
  webkit_settings_set_enable_developer_extras(settings, TRUE);

  gtk_window_set_child(GTK_WINDOW(window), GTK_WIDGET(web));

  // hook up some buttons
  WebKitUserContentManager *manager = webkit_web_view_get_user_content_manager(web);
  webkit_user_content_manager_register_script_message_handler(manager, "nativeApp", nullptr);
  g_signal_connect(manager, "script-message-received::nativeApp", G_CALLBACK(on_message), window);
  gtk_window_set_child(GTK_WINDOW(window), GTK_WIDGET(web));

  // librenderinghtml example: html=render_markdown(title,content)
  char *title = default_title;
  char *content = default_content;
  char *url = static_cast<char *>(user_data);
  char *htmlout;
  if (url != nullptr) {
    htmlout = render_markdown_file(url, url);
  } else {
    htmlout = render_markdown(title, content);
  }
  // show the output
  webkit_web_view_load_html(web, htmlout, nullptr);
  free_markdown(htmlout);
  gtk_widget_grab_focus(GTK_WIDGET(web));
  gtk_window_present(GTK_WINDOW(window));
}

int main(int argc, char **argv) {
  GtkApplication *app = gtk_application_new("arpa.home.aerth.markdownviewergtk", G_APPLICATION_DEFAULT_FLAGS);
  gchar *url = nullptr;
  if (argc == 2) {
    url = argv[1];
  }
  g_signal_connect(app, "activate", G_CALLBACK(activate), url);
  int status = g_application_run(G_APPLICATION(app), 0, NULL);

  g_object_unref(app);

  return status;
}
