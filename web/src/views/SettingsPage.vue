<template>
    <div class="settings-page">
        <div class="settings-container">
            <div class="settings-sidebar">
                <h1 class="settings-title">Settings</h1>
                <el-button
                    class="menu-btn"
                    :class="{ active: activeTab === 'password' }"
                    @click="activeTab = 'password'"
                >
                    Password
                </el-button>
                <el-button
                    class="menu-btn"
                    :class="{ active: activeTab === 'rename' }"
                    @click="activeTab = 'rename'"
                >
                    Rename
                </el-button>
                <el-button
                    class="menu-btn"
                    :class="{ active: activeTab === 'profile' }"
                    @click="activeTab = 'profile'"
                >
                    Profile
                </el-button>
            </div>

            <div class="settings-main">
                <div v-if="activeTab === 'password'" class="reset-section">
                    <h2 class="section-title">Reset Password</h2>

                    <div class="field-group">
                        <label class="field-label" for="old-password"
                            >Old Password</label
                        >
                        <el-input
                            id="old-password"
                            v-model="oldPassword"
                            type="password"
                            size="large"
                            placeholder="Enter old password"
                            show-password
                            class="field"
                            :class="{ 'is-error': oldPasswordError }"
                            @input="oldPasswordError = ''"
                            @keyup.enter="handleSubmit"
                        />
                        <p v-if="oldPasswordError" class="field-error">
                            {{ oldPasswordError }}
                        </p>
                    </div>

                    <div class="field-group">
                        <label class="field-label" for="new-password"
                            >New Password</label
                        >
                        <el-input
                            id="new-password"
                            v-model="newPassword"
                            type="password"
                            size="large"
                            placeholder="Enter new password"
                            show-password
                            class="field"
                            @keyup.enter="handleSubmit"
                        />
                    </div>

                    <div class="field-group">
                        <label class="field-label" for="confirm-password"
                            >Confirm New Password</label
                        >
                        <el-input
                            id="confirm-password"
                            v-model="confirmPassword"
                            type="password"
                            size="large"
                            placeholder="Confirm new password"
                            show-password
                            class="field"
                            @keyup.enter="handleSubmit"
                        />
                    </div>

                    <el-button
                        size="large"
                        class="submit-btn"
                        :loading="loading"
                        @click="handleSubmit"
                    >
                        Reset
                    </el-button>
                </div>

                <div v-if="activeTab === 'rename'" class="rename-section">
                    <h2 class="section-title">Rename</h2>

                    <div class="field-group">
                        <label class="field-label" for="new-username"
                            >New Username</label
                        >
                        <el-input
                            id="new-username"
                            v-model="newUsername"
                            type="text"
                            size="large"
                            placeholder="Enter new username"
                            maxlength="25"
                            class="field"
                            :class="{
                                'is-error':
                                    renameError ||
                                    isUsernameTooShort ||
                                    isUsernameStartsWithNumber,
                            }"
                            @input="renameError = ''"
                            @keyup.enter="handleRename"
                        />
                        <p class="field-hint">{{ newUsername.length }}/25</p>
                        <p v-if="renameError" class="field-error">
                            {{ renameError }}
                        </p>
                        <p
                            v-else-if="isUsernameStartsWithNumber"
                            class="field-error"
                        >
                            Username cannot start with a number
                        </p>
                        <p v-else-if="isUsernameTooShort" class="field-error">
                            Username must be at least 4 characters
                        </p>
                    </div>

                    <el-button
                        size="large"
                        class="submit-btn"
                        :loading="renameLoading"
                        @click="handleRename"
                    >
                        Save
                    </el-button>
                </div>

                <div v-if="activeTab === 'profile'" class="profile-section">
                    <h2 class="section-title">Edit Profile</h2>

                    <div class="field-group">
                        <label class="field-label" for="profile-links"
                            >Links</label
                        >
                        <div class="links-list">
                            <div
                                v-for="index in 5"
                                :key="index"
                                class="link-field"
                                :class="{ 'has-error': linkInvalid(index - 1) }"
                            >
                                <el-input
                                    :model-value="links[index - 1]"
                                    :placeholder="`https://link-${index}.com`"
                                    class="field"
                                    @update:model-value="
                                        onLinkInput(index - 1, $event)
                                    "
                                />
                                <span
                                    v-if="linkInvalid(index - 1)"
                                    class="link-hint"
                                >
                                    Must start with http(s)://
                                </span>
                            </div>
                        </div>
                    </div>

                    <div class="field-group">
                        <div class="field-label-row">
                            <label class="field-label" for="profile-bio"
                                >Bio</label
                            >
                            <span
                                class="char-count"
                                :class="{ over: bioOverLimit }"
                                >{{ bio.length }}/500</span
                            >
                        </div>
                        <el-input
                            id="profile-bio"
                            v-model="bio"
                            type="textarea"
                            :rows="8"
                            placeholder="Tell us a bit about yourself"
                            class="field"
                            :class="{ 'is-over': bioOverLimit }"
                        />
                    </div>

                    <div class="field-group email-visibility">
                        <el-checkbox v-model="showEmail">
                            Show email on my profile
                        </el-checkbox>
                        <p class="email-hint">
                            When enabled, your email is shown under your name
                            on your public profile. Off by default.
                        </p>
                    </div>

                    <el-button
                        size="large"
                        class="submit-btn"
                        :loading="saving"
                        :disabled="bioOverLimit || hasInvalidLink"
                        @click="saveProfile"
                    >
                        Save
                    </el-button>
                </div>
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from "vue";
import { useRouter } from "vue-router";
import { apiFetch, clearToken, getApiError, handleApiError } from "@/api";
import { notify } from "@/utils/message";
import { useUserStore } from "@/stores/user";

const router = useRouter();
const userStore = useUserStore();

const activeTab = ref<"password" | "rename" | "profile">("password");
const oldPassword = ref("");
const newPassword = ref("");
const confirmPassword = ref("");
const loading = ref(false);
const oldPasswordError = ref("");

const newUsername = ref("");
const renameLoading = ref(false);
const renameError = ref("");

// Edit profile form (moved from UserProfilePage)
const bio = ref("");
const links = ref<string[]>(Array(5).fill(""));
const showEmail = ref(false);
const saving = ref(false);

const isUsernameTooShort = computed(
    () => newUsername.value.length > 0 && newUsername.value.length < 4,
);

const isUsernameStartsWithNumber = computed(
    () => newUsername.value.length > 0 && /^\d/.test(newUsername.value),
);

const bioOverLimit = computed(() => bio.value.length > 500);

function isValidLink(value: string) {
    const trimmed = value.trim();
    return trimmed === "" || /^https?:\/\//i.test(trimmed);
}

function linkInvalid(index: number) {
    return !isValidLink(links.value[index] ?? "");
}

const hasInvalidLink = computed(() =>
    links.value.some((link) => !isValidLink(link)),
);

// Prefill the form with the current values each time the tab is opened.
watch(activeTab, (tab) => {
    if (tab === "profile") {
        bio.value = userStore.bio;
        links.value = Array(5).fill("");
        (userStore.links ?? []).slice(0, 5).forEach((link, index) => {
            links.value[index] = link;
        });
        showEmail.value = userStore.showEmail;
    }
});

function onLinkInput(index: number, value: string) {
    links.value[index] = value;
}

async function saveProfile() {
    if (bioOverLimit.value) {
        notify("error", "Bio must be 500 characters or fewer");
        return;
    }
    if (hasInvalidLink.value) {
        notify("error", "Links must start with http(s)://");
        return;
    }
    const filledLinks = links.value.map((link) => link.trim()).filter(Boolean);
    saving.value = true;
    try {
        const response = await apiFetch("/users/me/profile", {
            method: "PUT",
            body: JSON.stringify({
                bio: bio.value.trim(),
                links: filledLinks,
                show_email: showEmail.value,
            }),
        });
        if (!response.ok) {
            const message =
                (await getApiError(response)) ??
                `Failed to save profile (HTTP ${response.status})`;
            notify("error", message);
            return;
        }
        await userStore.fetchCurrentUser(true);
        notify("success", "Profile updated");
    } catch (error) {
        handleApiError(error, "Failed to save profile");
    } finally {
        saving.value = false;
    }
}

async function handleSubmit() {
    oldPasswordError.value = "";
    if (!oldPassword.value) {
        notify("warning", "Please enter your old password");
        return;
    }
    if (!newPassword.value) {
        notify("warning", "Please enter a new password");
        return;
    }
    if (newPassword.value !== confirmPassword.value) {
        notify("warning", "New passwords do not match");
        return;
    }

    loading.value = true;
    try {
        const response = await apiFetch("/users/reset", {
            method: "PATCH",
            body: JSON.stringify({
                old_password: oldPassword.value,
                new_password: newPassword.value,
            }),
        });
        if (!response.ok) {
            if (response.status === 401) {
                oldPasswordError.value = "Incorrect old password";
                return;
            }
            const json = await response.json().catch(() => null);
            const msg = json?.error || "Failed to update password";
            notify("error", msg);
            return;
        }

        notify("success", "Password updated");
        clearToken();
        router.push("/login");
    } catch (error) {
        handleApiError(error, "Failed to update password");
    } finally {
        loading.value = false;
    }
}

async function handleRename() {
    const name = newUsername.value.trim();
    if (!name) {
        notify("warning", "Please enter a new username");
        return;
    }
    if (name.length < 4) {
        renameError.value = "Username must be at least 4 characters";
        return;
    }
    if (/^\d/.test(name)) {
        renameError.value = "Username cannot start with a number";
        return;
    }
    if (name === userStore.username) {
        notify("warning", "New username is the same as the current username");
        return;
    }
    renameLoading.value = true;
    renameError.value = "";
    try {
        const response = await apiFetch("/users/rename", {
            method: "PATCH",
            body: JSON.stringify({ new_name: name }),
        });
        if (!response.ok) {
            if (response.status === 401) return;
            const message =
                (await getApiError(response)) ??
                `Failed to rename (HTTP ${response.status})`;
            notify("error", message);
            return;
        }
        await userStore.fetchCurrentUser(true);
        newUsername.value = "";
        notify("success", "Username updated");
    } catch (error) {
        handleApiError(error, "Failed to rename");
    } finally {
        renameLoading.value = false;
    }
}
</script>

<style scoped>
.settings-page {
    min-height: 100vh;
    padding: 32px 0 80px;
}

.settings-container {
    width: 75%;
    margin: 0 auto;
    display: flex;
    gap: 32px;
    padding: 0 24px;
}

.settings-sidebar {
    width: 180px;
    flex-shrink: 0;
}

.settings-title {
    margin: 0 0 24px;
    font-size: 22px;
    font-weight: 600;
    color: #e4e6e8;
}

.menu-btn {
    width: 100%;
    margin-left: 0 !important;
    text-align: left;
    justify-content: flex-start;
    background: transparent;
    color: #8c8c8c;
    border: 1px solid transparent;
    font-weight: 500;
}

.menu-btn:hover {
    color: #e4e6e8;
    background: transparent;
    border-color: transparent;
}

.menu-btn.active {
    color: #e4e6e8;
    background: #333;
    border-color: #333;
}

.settings-main {
    flex: 1;
    min-width: 0;
    display: flex;
    justify-content: center;
    align-items: flex-start;
}

.reset-section,
.rename-section,
.profile-section {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 24px;
    width: 100%;
    max-width: 620px;
}

.reset-section .field-group,
.rename-section .field-group {
    width: 320px;
}

.profile-section .field-group {
    width: 460px;
}

.profile-section .email-visibility {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    gap: 6px;
    padding: 14px 16px;
    border: 1px solid #262626;
    border-radius: 8px;
    background: #141414;
}

.profile-section .email-visibility :deep(.el-checkbox) {
    --el-checkbox-checked-bg-color: #6cbbf7;
    --el-checkbox-checked-input-border-color: #6cbbf7;
    --el-checkbox-checked-icon-color: #141414;
    --el-checkbox-input-border-color-hover: #6cbbf7;
}

.profile-section .email-visibility :deep(.el-checkbox__label) {
    color: #e4e6e8;
    font-size: 14px;
    font-weight: 500;
}

.profile-section .email-hint {
    margin: 0;
    font-size: 12px;
    line-height: 1.5;
    color: #8c8c8c;
}

.reset-section .submit-btn,
.rename-section .submit-btn {
    width: 320px;
}

.profile-section .submit-btn {
    width: 460px;
}

.links-list {
    display: flex;
    flex-direction: column;
    gap: 8px;
}

.link-field {
    display: flex;
    flex-direction: column;
}

.link-field.has-error :deep(.el-input__wrapper) {
    box-shadow: 0 0 0 1px #cf222e inset;
}

.link-field.has-error :deep(.el-input__wrapper.is-focus) {
    box-shadow:
        0 0 0 1px #cf222e inset,
        0 0 0 3px rgba(207, 34, 46, 0.2);
}

.link-hint {
    align-self: flex-end;
    margin-top: 2px;
    font-size: 12px;
    line-height: 1.4;
    color: #cf222e;
}

.field-label-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
}

.char-count {
    font-size: 12px;
    color: #8c8c8c;
}

.char-count.over {
    color: #cf222e;
    font-weight: 600;
}

.profile-section .field :deep(.el-textarea__inner) {
    background: transparent;
    box-shadow: 0 0 0 1px #333 inset;
    border-radius: 6px;
    color: #e4e6e8;
}

.profile-section .field :deep(.el-textarea__inner:focus) {
    box-shadow:
        0 0 0 1px #e4e6e8 inset,
        0 0 0 3px rgba(228, 230, 232, 0.1);
}

.profile-section .field :deep(.el-textarea__inner::placeholder) {
    color: #8c8c8c;
}

.field.is-over :deep(.el-textarea__inner) {
    border-color: #cf222e;
    box-shadow: 0 0 0 1px #cf222e inset;
}

.field.is-over :deep(.el-textarea__inner:focus) {
    border-color: #cf222e;
    box-shadow:
        0 0 0 1px #cf222e inset,
        0 0 0 3px rgba(207, 34, 46, 0.2);
}

.section-title {
    margin: 0 0 4px;
    font-size: 21px;
    font-weight: 600;
    color: #e4e6e8;
    text-align: center;
}

.field-group {
    display: flex;
    flex-direction: column;
    gap: 6px;
}

.field-label {
    font-size: 14px;
    font-weight: 600;
    color: #e4e6e8;
}

.field :deep(.el-input__wrapper) {
    background: transparent;
    box-shadow: 0 0 0 1px #333 inset;
    border-radius: 6px;
}

.field :deep(.el-input__wrapper.is-focus) {
    box-shadow:
        0 0 0 1px #e4e6e8 inset,
        0 0 0 3px rgba(228, 230, 232, 0.1);
}

.field :deep(.el-input__inner) {
    color: #e4e6e8;
}

.field :deep(.el-input__inner::placeholder) {
    color: #8c8c8c;
}

.field.is-error :deep(.el-input__wrapper),
.field.is-error :deep(.el-input__wrapper.is-focus) {
    box-shadow:
        0 0 0 1px #cf222e inset,
        0 0 0 3px rgba(207, 34, 46, 0.2);
}

.field-error {
    margin: 0;
    font-size: 13px;
    color: #cf222e;
}

.field-hint {
    margin: 0;
    font-size: 12px;
    color: #8c8c8c;
    text-align: right;
}

.submit-btn {
    width: 100%;
    background: #e4e6e8;
    color: #1a1a1a;
    border: 1px solid #e4e6e8;
    font-weight: 600;
    border-radius: 8px;
}

.submit-btn:hover {
    background: #ffffff;
    color: #1a1a1a;
    border-color: #ffffff;
}

.submit-btn.is-loading {
    background: #e4e6e8;
    border-color: #e4e6e8;
    color: #ffffff;
}
</style>
