<template>
  <div>
    <div v-if="validationError" class="alert alert-error mb-6">
      <span>{{ validationError }}</span>
    </div>

    <div class="card bg-base-100 shadow-md mb-6">
      <div class="card-body">
        <div class="flex justify-between items-center mb-4">
          <h2 class="card-title">Route Profiles</h2>
          <button class="btn btn-primary btn-sm" @click="showAddModal = true">
            Add Profile
          </button>
        </div>
        <div class="overflow-x-auto">
          <table class="table table-sm">
            <thead>
              <tr>
                <th>Name</th>
                <th>Label</th>
                <th>Description</th>
                <th>Downstream Alias</th>
                <th>Priority</th>
                <th>Enabled</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="(profile, name) in profiles"
                :key="name"
                :class="{ 'opacity-50': !profile.enabled }"
              >
                <td class="font-mono text-sm">{{ name }}</td>
                <td>{{ profile.label }}</td>
                <td class="max-w-xs truncate">{{ profile.description }}</td>
                <td class="font-mono text-sm">
                  {{ profile.downstream_alias }}
                </td>
                <td>
                  <span class="badge badge-outline badge-sm">{{
                    profile.priority
                  }}</span>
                </td>
                <td>
                  <input
                    type="checkbox"
                    class="toggle toggle-sm"
                    v-model="profile.enabled"
                    @change="dirty = true"
                    :disabled="name === 'route.default'"
                  />
                </td>
                <td>
                  <button
                    class="btn btn-ghost btn-xs"
                    @click="editProfile(name as string, profile)"
                  >
                    Edit
                  </button>
                  <button
                    class="btn btn-ghost btn-xs text-error"
                    @click="deleteProfile(name as string)"
                    v-if="name !== 'route.default'"
                  >
                    Delete
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <button class="btn btn-primary" @click="save" :disabled="!dirty">
      Save Changes
    </button>

    <dialog class="modal" :class="{ 'modal-open': showAddModal }">
      <div class="modal-box">
        <h3 class="font-bold text-lg">
          {{ editingName ? "Edit Profile" : "Add Profile" }}
        </h3>
        <div v-if="modalError" class="alert alert-error mt-2">
          <span>{{ modalError }}</span>
        </div>
        <div class="py-4 space-y-4">
          <div class="form-control">
            <label class="label"><span class="label-text">Name</span></label>
            <input
              class="input input-bordered"
              :class="{ 'input-error': formError.name }"
              v-model="form.name"
              :disabled="!!editingName"
              placeholder="route.custom"
              @input="clearFormError('name')"
            />
            <label class="label" v-if="formError.name">
              <span class="label-text-alt text-error">{{ formError.name }}</span>
            </label>
          </div>
          <div class="form-control">
            <label class="label"><span class="label-text">Label</span></label>
            <input
              class="input input-bordered"
              v-model="form.label"
              placeholder="Custom Route"
            />
          </div>
          <div class="form-control">
            <label class="label"
              ><span class="label-text">Description</span></label
            >
            <input class="input input-bordered" v-model="form.description" />
          </div>
          <div class="form-control">
            <label class="label"
              ><span class="label-text">Downstream Alias</span></label
            >
            <input
              class="input input-bordered"
              :class="{ 'input-error': formError.downstream_alias }"
              v-model="form.downstream_alias"
              placeholder="combo.custom"
              @input="clearFormError('downstream_alias')"
            />
            <label class="label" v-if="formError.downstream_alias">
              <span class="label-text-alt text-error">{{ formError.downstream_alias }}</span>
            </label>
          </div>
          <div class="form-control">
            <label class="label"
              ><span class="label-text">Priority</span></label
            >
            <select class="select select-bordered" v-model="form.priority">
              <option value="quality">Quality</option>
              <option value="balanced">Balanced</option>
              <option value="cost">Cost</option>
            </select>
          </div>
        </div>
        <div class="modal-action">
          <button class="btn btn-ghost" @click="closeModal">Cancel</button>
          <button class="btn btn-primary" @click="saveProfile">Save</button>
        </div>
      </div>
      <form method="dialog" class="modal-backdrop" @click="closeModal">
        <button>close</button>
      </form>
    </dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from "vue";
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

function load() {
  if (configStore.profiles) {
    profiles.value = JSON.parse(
      JSON.stringify(configStore.profiles.route_profiles),
    );
  }
  dirty.value = false;
}

function clearFormError(field: 'name' | 'downstream_alias') {
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

onMounted(load);
</script>
