import { Controller } from "@hotwired/stimulus";

class ToastButton extends Controller {
  toast() {
    document.dispatchEvent(
      new CustomEvent("basecoat:toast", {
        detail: {
          config: {
            category: "success",
            title: "Success",
            description: "A success toast called from the front-end.",
            cancel: {
              label: "Dismiss",
            },
          },
        },
      })
    );
  }
}

export default ToastButton;
