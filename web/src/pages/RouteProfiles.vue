<template>
  <div class="space-y-6 lg:space-y-8">
    <PageHeader
      eyebrow="Routing"
      title="Route Profiles"
      description="Manage reusable route profiles used by task routing and overrides."
    >
      <template #actions>
        <GradientButton @click="openCreateModal">Create Route Profile</GradientButton>
      </template>
    </PageHeader>

    <div v-if="validationError" class="rounded-2xl border border-rose-400/20 bg-rose-400/10 px-4 py-3 text-sm text-rose-100">
      {{ validationError }}
    </div>

    <div v-if="configStore.error" class="rounded-2xl border border-rose-400/20 bg-rose-400/10 px-4 py-3 text-sm text-rose-100">
      {{ configStore.error }}
    </div>

    <GlassCard>
      <div class="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
        <div>
          <h2 class="text-lg font-semibold text-white">Profile Library</h2>
          <p class="mt-1 text-sm text-slate-400">
            Each profile controls a target alias and priority mode for a route.
          </p>
        </div>

        <div class="flex flex-wrap gap-2 text-xs text-slate-400">
          <span class="rounded-full border border-white/10 bg-white/5 px-3 py-1">{{ orderedProfiles.length }} profiles</span>
          <span class="rounded-full border border-white/10 bg-white/5 px-3 py-1">{{ activeProfileCount }} active</span>
          <span class="rounded-full border border-white/10 bg-white/5 px-3 py-1">{{ inactiveProfileCount }} inactive</span>
        </div>
      </div>

      <div v-if="configStore.loading" class="mt-5 rounded-2xl border border-white/10 bg-white/5 p-6 text-sm text-slate-400">
        Loading route profiles...
      </div>

      <div v-else-if="orderedProfiles.length === 0" class="mt-5">
        <EmptyState
          title="No route profiles yet"
          description="Create your first profile to define a reusable route target and priority mode."
        >
          <template #icon>
            <span class="text-xl text-cyan-300">✦</span>
          </template>
          <template #actions>
            <GradientButton @click="openCreateModal">Create Route Profile</GradientButton>
          </template>
        </EmptyState>
      </div>

      <div v-else class="mt-5 grid gap-4 xl:grid-cols-2">
        <article
          v-for="{ name, profile } in orderedProfiles"
          :key="name"
          class="glass-card glow-border rounded-3xl p-5 sm:p-6 transition-all duration-200"
          :class="profile.enabled ? 'border-white/10' : 'opacity-85'"
        >
          <div class="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
            <div class="min-w-0 space-y-3">
              <div class="flex flex-wrap items-center gap-2">
                <h3 class="truncate text-lg font-semibold text-white">{{ name }}</h3>
                <StatusBadge :status="profile.enabled ? 'active' : 'inactive'" :label="profile.enabled ? 'active' : 'inactive'" />
                <span class="rounded-full border px-3 py-1 text-xs capitalize" :class="priorityClass(profile.priority)">
                  {{ priorityLabel(profile.priority) }}
                </span>
              </div>

              <p class="text-sm leading-6 text-slate-400">
                {{ profile.description || 'No description provided.' }}
              </p>

              <div class="flex flex-wrap gap-3 text-xs text-slate-400">
                <span class="rounded-full border border-white/10 bg-white/5 px-3 py-1">
                  Label: <span class="text-slate-100">{{ profile.label || 'Untitled' }}</span>
                </span>
                <span class="rounded-full border border-white/10 bg-white/5 px-3 py-1">
                  Target: <span class="font-mono text-cyan-200">{{ profile.downstream_alias || '-' }}</span>
                </span>
              </div>
            </div>

            <div class="flex items-center gap-2 sm:shrink-0">
              <button
                class="rounded-2xl border border-white/10 bg-white/5 px-3 py-2 text-xs font-medium text-slate-200 transition-all hover:border-cyan-400/25 hover:bg-white/10 disabled:opacity-50"
                @click="toggleProfile(name)"
                :disabled="name === 'route.default'"
              >
                {{ profile.enabled ? 'Disable' : 'Enable' }}
              </button>
              <button
                class="rounded-2xl border border-white/10 bg-white/5 px-3 py-2 text-xs font-medium text-slate-200 transition-all hover:border-cyan-400/25 hover:bg-white/10"
                @click="editProfile(name, profile)"
              >
                Edit
              </button>
              <button
                v-if="name !== 'route.default'"
                class="rounded-2xl border border-white/10 bg-white/5 px-3 py-2 text-xs font-medium text-rose-200 transition-all hover:border-rose-400/25 hover:bg-rose-400/10 hover:text-rose-100"
                @click="deleteProfile(name)"
              >
                Delete
              </button>
            </div>
          </div>
        </article>
      </div>
    </GlassCard>

    <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
      <div class="text-sm text-slate-400">
        Changes are staged locally until you save them.
      </div>
      <div class="flex gap-3">
        <GhostButton @click="load" :disabled="!dirty">Discard</GhostButton>
        <GradientButton @click="save" :disabled="!dirty">Save Changes</GradientButton>
      </div>
    </div>

    <dialog class="modal" :class="{ 'modal-open': showAddModal }">
      <div class="modal-box glass-card glow-border rounded-[1.75rem] bg-[rgba(8,12,22,0.95)] p-0 text-slate-100 shadow-[0_24px_80px_rgba(0,0,0,0.45)]">
        <div class="border-b border-white/10 px-6 py-5">
          <h3 class="text-lg font-semibold text-white">
            {{ editingName ? "Edit Profile" : "Add Profile" }}
          </h3>
          <p class="mt-1 text-sm text-slate-400">
            Define the profile name, target alias, and priority mode.
          </p>
        </div>

        <div v-if="modalError" class="mx-6 mt-5 rounded-2xl border border-rose-400/20 bg-rose-400/10 px-4 py-3 text-sm text-rose-100">
          {{ modalError }}
        </div>

        <div class="space-y-4 px-6 py-5">
          <FormField label="Name" :error="formError.name" required>
            <input
              class="w-full bg-transparent text-sm text-white outline-none placeholder:text-slate-500"
              v-model="form.name"
              :disabled="!!editingName"
              placeholder="route.custom"
              @input="clearFormError('name')"
            />
          </FormField>

          <FormField label="Label">
            <input
              class="w-full bg-transparent text-sm text-white outline-none placeholder:text-slate-500"
              v-model="form.label"
              placeholder="Custom Route"
            />
          </FormField>

          <FormField label="Description">
            <input
              class="w-full bg-transparent text-sm text-white outline-none placeholder:text-slate-500"
              v-model="form.description"
              placeholder="Describe when to use this route profile"
            />
          </FormField>

          <FormField label="Downstream Alias" :error="formError.downstream_alias" required>
            <input
              class="w-full bg-transparent text-sm text-white outline-none placeholder:text-slate-500"
              v-model="form.downstream_alias"
              placeholder="combo.custom"
              @input="clearFormError('downstream_alias')"
            />
          </FormField>

          <FormField label="Priority Mode">
            <select class="w-full bg-transparent text-sm text-white outline-none" v-model="form.priority">
              <option value="speed">Speed</option>
              <option value="balanced">Balanced</option>
              <option value="quality">Quality</option>
              <option value="cost">Cost</option>
            </select>
          </FormField>
        </div>

        <div class="flex items-center justify-end gap-3 border-t border-white/10 px-6 py-5">
          <GhostButton @click="closeModal">Cancel</GhostButton>
          <GradientButton @click="saveProfile">Save</GradientButton>
        </div>
      </div>
      <form method="dialog" class="modal-backdrop" @click="closeModal">
        <button>close</button>
      </form>
    </dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, onMounted } from "vue";
import { useConfigStore } from "../stores/config";
import { type RouteProfile } from "../api/client";
import GlassCard from "../components/ui/GlassCard.vue";
import PageHeader from "../components/ui/PageHeader.vue";
import GradientButton from "../components/ui/GradientButton.vue";
import GhostButton from "../components/ui/GhostButton.vue";
import EmptyState from "../components/ui/EmptyState.vue";
import StatusBadge from "../components/ui/StatusBadge.vue";
import FormField from "../components/ui/FormField.vue";

const configStore = useConfigStore();
const profiles = ref<Record<string, RouteProfile>>({});
const dirty = ref(false);
const showAddModal = ref(false);
const editingName = ref<string | null>(null);
const validationError = ref<string | null>(null);
const modalError = ref<string | null>(null);
const formError = ref<{ name?: string; downstream_alias?: string }>({});

const form = ref({
  name: "",
  label: "",
  description: "",
  downstream_alias: "",
  priority: "balanced" as string,
});

const orderedProfiles = computed(() =>
  Object.entries(profiles.value).map(([name, profile]) => ({ name, profile })),
);

const activeProfileCount = computed(() =>
  orderedProfiles.value.filter(({ profile }) => profile.enabled).length,
);

const inactiveProfileCount = computed(() =>
  orderedProfiles.value.filter(({ profile }) => !profile.enabled).length,
);

function load() {
  if (configStore.profiles) {
    profiles.value = JSON.parse(
      JSON.stringify(configStore.profiles.route_profiles),
    );
  }
  dirty.value = false;
}

function openCreateModal() {
  editingName.value = null;
  form.value = {
    name: "",
    label: "",
    description: "",
    downstream_alias: "",
    priority: "balanced",
  };
  formError.value = {};
  modalError.value = null;
  showAddModal.value = true;
}

function clearFormError(field: "name" | "downstream_alias") {
  if (formError.value[field]) {
    const newErrors = { ...formError.value };
    delete newErrors[field];
    formError.value = newErrors;
    modalError.value = null;
  }
}

function validateForm(): boolean {
  const errors: { name?: string; downstream_alias?: string } = {};

  if (!form.value.name.trim()) {
    errors.name = "Profile name is required";
  } else if (!editingName.value && profiles.value[form.value.name]) {
    errors.name = "Profile name already exists";
  } else if (
    editingName.value &&
    form.value.name !== editingName.value &&
    profiles.value[form.value.name]
  ) {
    errors.name = "Profile name already exists";
  }

  if (!form.value.downstream_alias.trim()) {
    errors.downstream_alias = "Downstream alias is required";
  }

  formError.value = errors;

  if (Object.keys(errors).length > 0) {
    modalError.value = "Please fix the errors below";
    return false;
  }

  return true;
}

function editProfile(name: string, profile: RouteProfile) {
  editingName.value = name;
  form.value = { name, ...profile };
  formError.value = {};
  modalError.value = null;
  showAddModal.value = true;
}

function deleteProfile(name: string) {
  delete profiles.value[name];
  dirty.value = true;
}

function toggleProfile(name: string) {
  if (name === "route.default") return;
  const profile = profiles.value[name];
  if (!profile) return;
  profile.enabled = !profile.enabled;
  dirty.value = true;
}

function closeModal() {
  showAddModal.value = false;
  editingName.value = null;
  form.value = {
    name: "",
    label: "",
    description: "",
    downstream_alias: "",
    priority: "balanced",
  };
  formError.value = {};
  modalError.value = null;
}

function saveProfile() {
  if (!validateForm()) return;

  profiles.value[form.value.name] = {
    label: form.value.label,
    description: form.value.description,
    downstream_alias: form.value.downstream_alias,
    priority: form.value.priority,
    enabled: true,
  };
  dirty.value = true;
  closeModal();
}

function validate(): string | null {
  for (const [name, profile] of Object.entries(profiles.value)) {
    if (!profile.downstream_alias || !profile.downstream_alias.trim()) {
      return `Profile "${name}" has empty downstream alias`;
    }
  }
  return null;
}

async function save() {
  const error = validate();
  if (error) {
    validationError.value = error;
    return;
  }
  try {
    await configStore.saveProfiles({ route_profiles: profiles.value });
    validationError.value = null;
    dirty.value = false;
  } catch (e: any) {
    validationError.value = e.message;
  }
}

function priorityLabel(priority: string) {
  return priority || "balanced";
}

function priorityClass(priority: string) {
  switch (priority) {
    case "speed":
      return "border-cyan-400/20 bg-cyan-400/10 text-cyan-200";
    case "quality":
      return "border-violet-400/20 bg-violet-400/10 text-violet-200";
    case "cost":
      return "border-amber-400/20 bg-amber-400/10 text-amber-200";
    default:
      return "border-blue-400/20 bg-blue-400/10 text-blue-200";
  }
}

onMounted(load);
</script>
