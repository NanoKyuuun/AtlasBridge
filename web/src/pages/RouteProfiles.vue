<template>
  <div>
    <!-- Error Banner -->
    <div v-if="validationError" class="mb-4 p-4 rounded-lg border border-[var(--red)] bg-[rgba(248,113,113,.1)] text-[var(--red)] text-[13px]">
      {{ validationError }}
    </div>

    <!-- Header Bar -->
    <div class="flex items-center justify-between mb-5">
      <div>
        <div class="text-[14px] font-semibold">Route Profiles</div>
        <div class="text-[11.5px] text-[var(--text-mute)]">Kelola abstraksi routing untuk 9Router</div>
      </div>
      <div class="flex gap-2">
        <button v-if="dirty" class="btn btn-primary" @click="save">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M19 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11l5 5v11a2 2 0 0 1-2 2z"/><polyline points="17 21 17 13 7 13 7 21"/><polyline points="7 3 7 8 15 8"/></svg>
          Save Changes
        </button>
        <button class="btn btn-primary" @click="openCreateModal">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
          New Profile
        </button>
      </div>
    </div>

    <!-- Loading -->
    <div v-if="configStore.loading" class="card p-8 text-center text-[var(--text-mute)] text-[13px]">
      Loading profiles...
    </div>

    <!-- Profile Cards Grid -->
    <div v-else class="grid grid-cols-3 gap-4">
      <!-- Existing Profiles -->
      <div v-for="{ name, profile } in orderedProfiles" :key="name" class="profile-card">
        <div class="flex items-start justify-between mb-3">
          <div class="w-10 h-10 rounded-xl flex items-center justify-center" style="background: rgba(124, 92, 255, .15);">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="#a78bfa" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="16 18 22 12 16 6"/><polyline points="8 6 2 12 8 18"/></svg>
          </div>
          <div class="flex items-center gap-2">
            <span class="badge" :class="profile.enabled ? 'badge-green' : 'badge-gray'">
              {{ profile.enabled ? 'Active' : 'Inactive' }}
            </span>
          </div>
        </div>
        <div class="font-semibold text-[14px] mb-1">{{ name }}</div>
        <div class="text-[12px] text-[var(--text-dim)] mb-3">{{ profile.description || profile.label || 'No description' }}</div>
        <div class="flex items-center gap-2 mb-3">
          <span class="code-tag">{{ profile.downstream_alias }}</span>
          <span class="badge" :class="priorityBadgeClass(profile.priority)">{{ priorityLabel(profile.priority) }}</span>
        </div>
        <div class="flex items-center justify-between text-[11.5px]">
          <button
            class="text-[var(--text-mute)] hover:text-[var(--yellow)] transition-colors text-[12px]"
            @click="toggleProfile(name)"
            :disabled="name === 'route.default'"
          >
            {{ profile.enabled ? 'Disable' : 'Enable' }}
          </button>
          <div class="flex gap-2">
            <button class="text-[var(--accent)] hover:underline text-[12px]" @click="editProfile(name, profile)">Edit</button>
            <button
              v-if="name !== 'route.default'"
              class="text-[var(--red)] hover:underline text-[12px]"
              @click="deleteProfile(name)"
            >Delete</button>
          </div>
        </div>
      </div>

      <!-- Create New Profile Card (dashed) -->
      <div class="profile-card" style="border-style: dashed; opacity: .7;" @click="openCreateModal">
        <div class="flex items-start justify-between mb-3">
          <div class="w-10 h-10 rounded-xl flex items-center justify-center border border-dashed border-[var(--border)]">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
          </div>
        </div>
        <div class="font-semibold text-[14px] mb-1 text-[var(--text-dim)]">Create New Profile</div>
        <div class="text-[12px] text-[var(--text-mute)]">Define custom route profile for 9Router</div>
      </div>
    </div>

    <!-- Modal -->
    <div v-if="showAddModal" class="modal-backdrop" @click.self="closeModal">
      <div class="modal">
        <div class="flex items-center justify-between mb-5">
          <div>
            <div class="text-[16px] font-semibold">{{ editingName ? 'Edit Profile' : 'Create Route Profile' }}</div>
            <div class="text-[11.5px] text-[var(--text-mute)]">Definisikan route profile untuk 9Router</div>
          </div>
          <button class="btn btn-ghost" @click="closeModal">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
          </button>
        </div>

        <div v-if="modalError" class="mb-4 p-3 rounded-lg border border-[var(--red)] bg-[rgba(248,113,113,.1)] text-[var(--red)] text-[12px]">
          {{ modalError }}
        </div>

        <div class="space-y-3">
          <div>
            <label class="text-[12px] text-[var(--text-mute)] mb-1.5 block">Profile Name</label>
            <input
              type="text" class="input"
              placeholder="e.g. route.custom_reasoning"
              v-model="form.name"
              @input="clearFormError('name')"
              :disabled="!!editingName"
            >
            <div v-if="formError.name" class="mt-1 text-[11px] text-[var(--red)]">{{ formError.name }}</div>
          </div>
          <div>
            <label class="text-[12px] text-[var(--text-mute)] mb-1.5 block">Label (Display Name)</label>
            <input type="text" class="input" placeholder="e.g. Backend Engineering" v-model="form.label">
          </div>
          <div>
            <label class="text-[12px] text-[var(--text-mute)] mb-1.5 block">Description</label>
            <textarea class="textarea" rows="2" placeholder="Describe when this route should be used..." v-model="form.description"></textarea>
          </div>
          <div>
            <label class="text-[12px] text-[var(--text-mute)] mb-1.5 block">Downstream Alias</label>
            <input
              type="text" class="input mono"
              placeholder="e.g. x-9router-profile: reasoning-v2"
              v-model="form.downstream_alias"
              @input="clearFormError('downstream_alias')"
            >
            <div v-if="formError.downstream_alias" class="mt-1 text-[11px] text-[var(--red)]">{{ formError.downstream_alias }}</div>
          </div>
          <div>
            <label class="text-[12px] text-[var(--text-mute)] mb-1.5 block">Priority Mode</label>
            <div class="grid grid-cols-4 gap-2">
              <label v-for="mode in ['speed', 'quality', 'cost', 'balanced']" :key="mode"
                class="card-soft p-2.5 text-center cursor-pointer border-2 transition-all"
                :class="form.priority === mode ? 'border-[var(--accent)]' : 'border-transparent hover:border-[var(--accent)]'"
              >
                <input type="radio" name="priority" class="hidden" :value="mode" v-model="form.priority">
                <div class="text-[12px] font-medium capitalize">{{ mode }}</div>
              </label>
            </div>
          </div>
        </div>

        <div class="flex justify-end gap-2 mt-5">
          <button class="btn btn-ghost" @click="closeModal">Cancel</button>
          <button class="btn btn-primary" @click="saveProfile">{{ editingName ? 'Update' : 'Create Profile' }}</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, onMounted } from "vue";
import { useConfigStore } from "../stores/config";
import { type RouteProfile } from "../api/client";

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

function load() {
  if (configStore.profiles) {
    profiles.value = JSON.parse(JSON.stringify(configStore.profiles.route_profiles));
  }
  dirty.value = false;
}

function openCreateModal() {
  editingName.value = null;
  form.value = { name: "", label: "", description: "", downstream_alias: "", priority: "balanced" };
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
  } else if (editingName.value && form.value.name !== editingName.value && profiles.value[form.value.name]) {
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
  form.value = { name: "", label: "", description: "", downstream_alias: "", priority: "balanced" };
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
  if (error) { validationError.value = error; return; }
  try {
    await configStore.saveProfiles({ route_profiles: profiles.value });
    validationError.value = null;
    dirty.value = false;
  } catch (e: any) {
    validationError.value = e.message;
  }
}

function priorityLabel(priority: string) { return priority || "balanced"; }

function priorityBadgeClass(priority: string) {
  switch (priority) {
    case "speed": return "badge-blue";
    case "quality": return "badge-purple";
    case "cost": return "badge-yellow";
    default: return "badge-gray";
  }
}

onMounted(load);
</script>
