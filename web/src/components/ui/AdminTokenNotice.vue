<template>
  <div class="rounded-3xl border border-amber-400/20 bg-amber-400/10 p-5 sm:p-6">
    <div class="flex items-start justify-between gap-4">
      <div class="min-w-0 flex-1">
        <div class="flex items-center gap-2 text-sm font-semibold text-amber-200">
          <svg class="h-4 w-4 shrink-0" viewBox="0 0 16 16" fill="currentColor">
            <path d="M8 1a7 7 0 1 0 0 14A7 7 0 0 0 8 1ZM7 5a1 1 0 1 1 2 0v2a1 1 0 1 1-2 0V5Zm1 6a1 1 0 1 1 0-2 1 1 0 0 1 0 2Z"/>
          </svg>
          Admin Token Generated
        </div>
        <p class="mt-1 text-xs leading-5 text-amber-100/70">
          Save this token now. You will not be able to see it again. Store it in a safe place.
        </p>
        <div class="mt-3 flex items-center gap-2">
          <code class="select-all break-all rounded-xl border border-amber-400/15 bg-amber-950/40 px-3 py-2 font-mono text-sm text-amber-100">{{ token }}</code>
          <button
            type="button"
            class="inline-flex shrink-0 items-center justify-center gap-1.5 rounded-xl border border-amber-400/20 bg-amber-400/10 px-3 py-2 text-xs font-medium text-amber-200 transition-all duration-200 hover:border-amber-400/30 hover:bg-amber-400/15 active:scale-[0.97]"
            @click="copy"
          >
            <svg v-if="!copied" class="h-3.5 w-3.5" viewBox="0 0 16 16" fill="currentColor">
              <path d="M4 2a2 2 0 0 1 2-2h5.586A2 2 0 0 1 13 .586l1.414 1.414A2 2 0 0 1 15 3.414V10a2 2 0 0 1-2 2H9a2 2 0 0 1-2-2V4a2 2 0 0 0-2-2H4Z"/>
              <path d="M2 5a2 2 0 0 0-2 2v6a2 2 0 0 0 2 2h5.586A2 2 0 0 0 9 13.586l.414-.414A1 1 0 0 0 9 13H2a1 1 0 0 1-1-1V6a1 1 0 0 1 1-1h4.586A2 2 0 0 1 7 4.586l.414-.414A1 1 0 0 0 7 4H2Z"/>
            </svg>
            <svg v-else class="h-3.5 w-3.5" viewBox="0 0 16 16" fill="currentColor">
              <path d="M13.485 1.929a1 1 0 0 1 1.414 1.414l-9 9a1 1 0 0 1-1.414 0l-4-4a1 1 0 0 1 1.414-1.414L5 10.586l8.485-8.485Z"/>
            </svg>
            {{ copied ? "Copied" : "Copy" }}
          </button>
        </div>
      </div>
      <button
        type="button"
        class="shrink-0 rounded-xl p-1.5 text-amber-200/50 transition-all duration-200 hover:bg-amber-400/10 hover:text-amber-200"
        @click="$emit('dismiss')"
        aria-label="Dismiss token notice"
      >
        <svg class="h-4 w-4" viewBox="0 0 16 16" fill="currentColor">
          <path d="M4.646 4.646a.5.5 0 0 1 .708 0L8 7.293l2.646-2.647a.5.5 0 0 1 .708.708L8.707 8l2.647 2.646a.5.5 0 0 1-.708.708L8 8.707l-2.646 2.647a.5.5 0 0 1-.708-.708L7.293 8 4.646 5.354a.5.5 0 0 1 0-.708Z"/>
        </svg>
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from "vue";

const props = defineProps<{
  token: string;
}>();

defineEmits<{
  dismiss: [];
}>();

const copied = ref(false);

async function copy() {
  try {
    await navigator.clipboard.writeText(props.token);
    copied.value = true;
    setTimeout(() => (copied.value = false), 2000);
  } catch {
    // clipboard unavailable
  }
}
</script>
