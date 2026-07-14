import { ref } from "vue";

export type ToastType = "success" | "warning" | "error" | "info";

export interface ToastItem {
  id: number;
  message: string;
  type: ToastType;
}

const toasts = ref<ToastItem[]>([]);
let nextId = 0;

export function useToast() {
  function showToast(message: string, type: ToastType = "info") {
    const id = nextId++;
    toasts.value.push({ id, message, type });
    setTimeout(() => {
      toasts.value = toasts.value.filter((t) => t.id !== id);
    }, 3000);
  }

  function removeToast(id: number) {
    toasts.value = toasts.value.filter((t) => t.id !== id);
  }

  return { toasts, showToast, removeToast };
}
