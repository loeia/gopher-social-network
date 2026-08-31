<template>
    <div class="user-profile-page">
        <div class="profile-card" v-loading="loading">
            <template v-if="notFound">
                <h1 class="not-found-title">User not found</h1>
                <p class="not-found-hint">
                    This user may have been removed or the link is incorrect.
                </p>
            </template>
            <template v-else>
                <div class="profile-header">
                    <div class="avatar-wrapper" :class="{ own: isOwnProfile }">
                        <UserAvatar
                            :user-id="userId"
                            :username="user.username"
                            :size="112"
                        />
                        <label v-if="isOwnProfile" class="avatar-overlay">
                            <input
                                ref="avatarInputRef"
                                type="file"
                                accept="image/*"
                                class="hidden-input"
                                @change="onAvatarFileChange"
                            />
                            <svg
                                class="camera-icon"
                                viewBox="0 0 24 24"
                                fill="none"
                                stroke="currentColor"
                                stroke-width="2"
                                stroke-linecap="round"
                                stroke-linejoin="round"
                            >
                                <path
                                    d="M23 19a2 2 0 0 1-2 2H3a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h4l2-3h6l2 3h4a2 2 0 0 1 2 2z"
                                />
                                <circle cx="12" cy="13" r="4" />
                            </svg>
                        </label>
                        <button
                            v-if="isOwnProfile"
                            class="avatar-delete-btn"
                            title="Remove avatar"
                            :disabled="avatarDeleting"
                            @click.stop="deleteAvatar"
                        >
                            &minus;
                        </button>
                    </div>

                    <h1 class="username">{{ user.username }}</h1>
                    <p v-if="user.show_email && user.email" class="email">
                        {{ user.email }}
                    </p>
                    <span class="handle"
                        >Joined {{ formatDate(user.created_at) }}</span
                    >

                    <p v-if="profileBio" class="bio">{{ profileBio }}</p>

                    <div class="meta-row">
                        <span
                            v-for="(link, index) in profileLinks"
                            :key="index"
                            class="meta-item"
                        >
                            <svg
                                class="info-icon"
                                viewBox="0 0 24 24"
                                fill="none"
                                stroke="currentColor"
                                stroke-width="2"
                                stroke-linecap="round"
                                stroke-linejoin="round"
                            >
                                <path
                                    d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71"
                                />
                                <path
                                    d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71"
                                />
                            </svg>
                            <a
                                :href="link"
                                class="website"
                                target="_blank"
                                rel="noopener noreferrer"
                            >
                                {{ link }}
                            </a>
                        </span>
                    </div>

                    <div class="stats-row">
                        <span class="stat clickable" @click="goToTab('posts')">
                            <strong>{{ user.posts_count }}</strong>
                            <span class="stat-label">Posts</span>
                        </span>
                        <span
                            class="stat clickable"
                            @click="goToTab('replies')"
                        >
                            <strong>{{ user.replies_count }}</strong>
                            <span class="stat-label">Replies</span>
                        </span>
                        <span class="stat clickable" @click="goToTab('likes')">
                            <strong>{{ user.likes_count }}</strong>
                            <span class="stat-label">Likes</span>
                        </span>
                        <span class="stat clickable" @click="goToFollowers()">
                            <strong>{{ user.followers_count }}</strong>
                            <span class="stat-label">Followers</span>
                        </span>
                        <span class="stat clickable" @click="goToFollowing()">
                            <strong>{{ user.following_count }}</strong>
                            <span class="stat-label">Following</span>
                        </span>
                    </div>
                </div>
            </template>
        </div>

        <div v-if="!notFound && !loading" class="profile-tabs">
            <div class="tabs-bar" ref="tabsBarRef">
                <button
                    v-for="tab in tabs"
                    :key="tab.key"
                    class="tab-btn"
                    :class="{ active: activeTab === tab.key }"
                    @click="activeTab = tab.key"
                >
                    {{ tab.label }}
                </button>
            </div>

            <div class="tabs-content">
                <div v-show="activeTab === 'posts'" class="tab-panel">
                    <div v-loading="postsLoading" class="feed">
                        <div
                            v-for="post in userPosts"
                            :key="post.id"
                            class="topic-row"
                            :class="{ 'new-item': postsHighlightId === post.id }"
                            @click="openPost(post.id)"
                        >
                            <div class="topic-top">
                                <h2 class="topic-title">{{ post.title }}</h2>
                                <div class="topic-stats">
                                    <span
                                        class="topic-stat"
                                        :title="`${post.comment_count} comments`"
                                    >
                                        <svg
                                            class="stat-icon"
                                            viewBox="0 0 24 24"
                                            aria-hidden="true"
                                        >
                                            <path
                                                d="M20 2H4a2 2 0 0 0-2 2v18l4-4h14a2 2 0 0 0 2-2V4a2 2 0 0 0-2-2z"
                                            />
                                        </svg>
                                        {{ post.comment_count }}
                                    </span>
                                    <span
                                        class="topic-stat"
                                        :title="`${post.like_count} likes`"
                                    >
                                        <svg
                                            class="stat-icon like-icon"
                                            viewBox="0 0 24 24"
                                            aria-hidden="true"
                                        >
                                            <path
                                                d="M12 21.35l-1.45-1.32C5.4 15.36 2 12.28 2 8.5 2 5.42 4.42 3 7.5 3c1.74 0 3.41.81 4.5 2.09C13.09 3.81 14.76 3 16.5 3 19.58 3 22 5.42 22 8.5c0 3.78-3.4 6.86-8.55 11.54L12 21.35z"
                                            />
                                        </svg>
                                        {{ post.like_count }}
                                    </span>
                                    <span
                                        class="topic-stat"
                                        :title="`${post.view_count} views`"
                                    >
                                        <svg
                                            class="stat-icon"
                                            viewBox="0 0 24 24"
                                            aria-hidden="true"
                                        >
                                            <path
                                                d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"
                                            />
                                            <circle cx="12" cy="12" r="3" />
                                        </svg>
                                        {{ post.view_count }}
                                    </span>
                                </div>
                            </div>
                            <div class="topic-meta">
                                <span class="topic-date">{{
                                    formatDate(post.created_at)
                                }}</span>
                                <template v-if="post.tags && post.tags.length">
                                    <span class="meta-dot">&middot;</span>
                                    <span
                                        v-for="tag in post.tags"
                                        :key="tag"
                                        class="topic-tag"
                                        >{{ tag }}</span
                                    >
                                </template>
                            </div>
                        </div>
                        <div
                            v-if="!postsLoading && userPosts.length === 0"
                            class="empty-hint"
                        >
                            No posts yet.
                        </div>
                        <div v-if="postsLoadingMore" class="loading-more">
                            Loading...
                        </div>
                    </div>
                </div>

                <div v-show="activeTab === 'replies'" class="tab-panel">
                    <div v-loading="repliesLoading" class="feed">
                        <div
                            v-for="reply in userReplies"
                            :key="reply.id"
                            class="topic-row reply-row"
                            :class="{
                                'new-item': repliesHighlightId === reply.id,
                            }"
                            @click="openPost(reply.post_id)"
                        >
                            <div class="reply-context">
                                <svg
                                    class="reply-icon"
                                    viewBox="0 0 24 24"
                                    aria-hidden="true"
                                >
                                    <path
                                        d="M9 17H7v-7h2v7zm4 0h-2V7h2v10zm4 0h-2v-4h2v4z"
                                    />
                                </svg>
                                <span class="reply-hint"
                                    >replied to a post</span
                                >
                            </div>
                            <p class="reply-content">{{ reply.content }}</p>
                            <div class="topic-meta">
                                <span class="topic-date">{{
                                    formatDate(reply.created_at)
                                }}</span>
                            </div>
                        </div>
                        <div
                            v-if="!repliesLoading && userReplies.length === 0"
                            class="empty-hint"
                        >
                            No replies yet.
                        </div>
                        <div v-if="repliesLoadingMore" class="loading-more">
                            Loading...
                        </div>
                    </div>
                </div>

                <div v-show="activeTab === 'likes'" class="tab-panel">
                    <div v-loading="likesLoading" class="feed">
                        <div
                            v-for="post in userLikedPosts"
                            :key="post.post_id"
                            class="topic-row"
                            :class="{
                                'new-item': likesHighlightId === post.post_id,
                            }"
                            @click="openPost(post.post_id)"
                        >
                            <div class="topic-top">
                                <h2 class="topic-title">{{ post.title }}</h2>
                                <div class="topic-stats">
                                    <span
                                        class="topic-stat"
                                        :title="`${post.comment_count} comments`"
                                    >
                                        <svg
                                            class="stat-icon"
                                            viewBox="0 0 24 24"
                                            aria-hidden="true"
                                        >
                                            <path
                                                d="M20 2H4a2 2 0 0 0-2 2v18l4-4h14a2 2 0 0 0 2-2V4a2 2 0 0 0-2-2z"
                                            />
                                        </svg>
                                        {{ post.comment_count ?? 0 }}
                                    </span>
                                    <span
                                        class="topic-stat"
                                        :title="`${post.like_count} likes`"
                                    >
                                        <svg
                                            class="stat-icon like-icon"
                                            viewBox="0 0 24 24"
                                            aria-hidden="true"
                                        >
                                            <path
                                                d="M12 21.35l-1.45-1.32C5.4 15.36 2 12.28 2 8.5 2 5.42 4.42 3 7.5 3c1.74 0 3.41.81 4.5 2.09C13.09 3.81 14.76 3 16.5 3 19.58 3 22 5.42 22 8.5c0 3.78-3.4 6.86-8.55 11.54L12 21.35z"
                                            />
                                        </svg>
                                        {{ post.like_count ?? 0 }}
                                    </span>
                                </div>
                            </div>
                            <div class="topic-bottom">
                                <span class="topic-author">{{
                                    post.author
                                }}</span>
                                <template v-if="post.tags && post.tags.length">
                                    <span class="meta-dot">&middot;</span>
                                    <span
                                        v-for="tag in post.tags"
                                        :key="tag"
                                        class="topic-tag"
                                        >{{ tag }}</span
                                    >
                                </template>
                            </div>
                        </div>
                        <div
                            v-if="!likesLoading && userLikedPosts.length === 0"
                            class="empty-hint"
                        >
                            No liked posts yet.
                        </div>
                        <div v-if="likesLoadingMore" class="loading-more">
                            Loading...
                        </div>
                    </div>
                </div>
            </div>
        </div>

        <AvatarCropDialog
            :visible="cropVisible"
            :src="cropSrc"
            @confirm="uploadAvatar"
            @close="cropVisible = false"
        />
    </div>
</template>

<script setup lang="ts">
import {
    computed,
    nextTick,
    onActivated,
    onBeforeUnmount,
    onDeactivated,
    ref,
    watch,
} from "vue";
import { useRoute, useRouter } from "vue-router";
import { apiFetch, getApiError, getCurrentUserId, handleApiError } from "@/api";
import { notify } from "@/utils/message";
import UserAvatar from "@/components/UserAvatar.vue";
import AvatarCropDialog from "@/components/AvatarCropDialog.vue";
import { useUserStore } from "@/stores/user";

defineOptions({ name: "UserProfilePage" });

interface UserProfile {
    id: number;
    username: string;
    email: string;
    show_email: boolean;
    created_at: string;
    bio: string;
    links: string[];
    followers_count: number;
    following_count: number;
    posts_count: number;
    likes_count: number;
    replies_count: number;
}

interface UserPost {
    id: number;
    title: string;
    tags: string[];
    comment_count: number;
    like_count: number;
    view_count: number;
    created_at: string;
}

interface UserReply {
    id: number;
    post_id: number;
    content: string;
    created_at: string;
}

interface LikedPost {
    post_id: number;
    author: string;
    title: string;
    tags: string[];
    comment_count?: number;
    like_count?: number;
    created_at: string;
}

const route = useRoute();
const router = useRouter();
const userStore = useUserStore();

// ID of the user whose data is currently loaded. Used to avoid re-fetching when
// returning to the same profile (keep-alive restores the cached DOM, so the
// router's scrollBehavior can restore the exact scroll position, same as Home).
let loadedUserId = 0;

const loading = ref(false);
const notFound = ref(false);
const user = ref<UserProfile>({
    id: 0,
    username: "",
    email: "",
    show_email: false,
    created_at: "",
    bio: "",
    links: [],
    followers_count: 0,
    following_count: 0,
    posts_count: 0,
    likes_count: 0,
    replies_count: 0,
});

const avatarInputRef = ref<HTMLInputElement | null>(null);
const cropVisible = ref(false);
const cropSrc = ref("");
const avatarDeleting = ref(false);

const activeTab = ref<"posts" | "replies" | "likes">("posts");
const tabs = [
    { key: "posts" as const, label: "Posts" },
    { key: "replies" as const, label: "Replies" },
    { key: "likes" as const, label: "Likes" },
];

const userPosts = ref<UserPost[]>([]);
const postsLoading = ref(false);
const postsOffset = ref(0);
const postsHasMore = ref(true);
const postsHighlightId = ref<number | null>(null);
const userReplies = ref<UserReply[]>([]);
const repliesLoading = ref(false);
const repliesOffset = ref(0);
const repliesHasMore = ref(true);
const repliesHighlightId = ref<number | null>(null);
const userLikedPosts = ref<LikedPost[]>([]);
const likesLoading = ref(false);
const likesOffset = ref(0);
const likesHasMore = ref(true);
const likesHighlightId = ref<number | null>(null);

const userId = computed(() => Number(route.params.userId));
const isOwnProfile = computed(() => getCurrentUserId() === userId.value);

const tabsBarRef = ref<HTMLElement | null>(null);

// When viewing the own profile, prefer the user store so that bio / links
// changes made in Settings show up immediately even with keep-alive.
const profileBio = computed(() => {
    if (isOwnProfile.value && userStore.loaded) return userStore.bio;
    return user.value.bio;
});

const profileLinks = computed(() => {
    if (isOwnProfile.value && userStore.loaded) return userStore.links;
    return user.value.links;
});

function formatDate(value?: string) {
    if (!value) return "";
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) return value;
    const y = date.getFullYear();
    const m = String(date.getMonth() + 1).padStart(2, "0");
    const d = String(date.getDate()).padStart(2, "0");
    return `${y}/${m}/${d}`;
}

function goToTab(tab: "posts" | "replies" | "likes") {
    activeTab.value = tab;
    nextTick(() => {
        const bar = tabsBarRef.value;
        if (!bar) return;
        // Scroll so the tab bar sits right below the sticky navbar.
        const navbar = document.querySelector(".navbar");
        const navHeight = navbar ? navbar.getBoundingClientRect().height : 0;
        const top =
            bar.getBoundingClientRect().top + window.scrollY - navHeight;
        window.scrollTo({ top: Math.max(0, top), behavior: "smooth" });
    });
}

function goToFollowers() {
    router.push({
        name: "Followers",
        params: { userId: String(userId.value) },
    });
}

function goToFollowing() {
    router.push({
        name: "Following",
        params: { userId: String(userId.value) },
    });
}

function onAvatarFileChange(e: Event) {
    const file = (e.target as HTMLInputElement).files?.[0];
    if (!file) return;
    cropSrc.value = URL.createObjectURL(file);
    cropVisible.value = true;
    if (avatarInputRef.value) avatarInputRef.value.value = "";
}

async function uploadAvatar(blob: Blob) {
    try {
        const formData = new FormData();
        formData.append("avatar", blob, "avatar.jpg");
        const response = await apiFetch("/users/me/avatar", {
            method: "PUT",
            body: formData,
        });
        if (!response.ok) {
            const message =
                (await getApiError(response)) ??
                `Failed to upload avatar (HTTP ${response.status})`;
            notify("error", message);
            return;
        }
        userStore.bumpAvatarVersion();
        cropVisible.value = false;
        notify("success", "Avatar updated");
    } catch (error) {
        handleApiError(error, "Failed to upload avatar");
    }
}

async function deleteAvatar() {
    if (avatarDeleting.value) return;
    avatarDeleting.value = true;
    try {
        const response = await apiFetch("/users/me/avatar", {
            method: "DELETE",
        });
        if (!response.ok) {
            const message =
                (await getApiError(response)) ??
                `Failed to delete avatar (HTTP ${response.status})`;
            notify("error", message);
            return;
        }
        userStore.bumpAvatarVersion();
        notify("success", "Avatar removed");
    } catch (error) {
        handleApiError(error, "Failed to delete avatar");
    } finally {
        avatarDeleting.value = false;
    }
}

async function load() {
    const id = userId.value;
    if (!id) return;
    loadedUserId = id;
    loading.value = true;
    notFound.value = false;
    // A fresh load must never carry over a leftover highlight from a previous
    // visit (e.g. a load-more flash interrupted by navigation). Only items
    // fetched by a later load-more get highlighted.
    postsHighlightId.value = null;
    repliesHighlightId.value = null;
    likesHighlightId.value = null;
    user.value = {
        id,
        username: "",
        email: "",
        show_email: false,
        created_at: "",
        bio: "",
        links: [],
        followers_count: 0,
        following_count: 0,
        posts_count: 0,
        likes_count: 0,
        replies_count: 0,
    };
    try {
        const response = await apiFetch(`/users/${id}`);
        if (!response.ok) {
            if (response.status === 404) {
                notFound.value = true;
                return;
            }
            throw new Error(`HTTP ${response.status}`);
        }
        const json = await response.json();
        const data = json.data ?? json;
        user.value = {
            id: Number(data.id),
            username: data.username ?? "",
            email: data.email ?? "",
            show_email: !!data.show_email,
            created_at: data.created_at ?? "",
            bio: data.bio ?? "",
            links: Array.isArray(data.links) ? data.links : [],
            followers_count: Number(data.followers_count ?? 0),
            following_count: Number(data.following_count ?? 0),
            posts_count: Number(data.posts_count ?? 0),
            likes_count: Number(data.likes_count ?? 0),
            replies_count: Number(data.replies_count ?? 0),
        };

        postsLoading.value = true;
        repliesLoading.value = true;
        likesLoading.value = true;
        const [postsRes, repliesRes, likesRes] = await Promise.all([
            apiFetch(`/users/${id}/posts?limit=20&offset=0&sort=desc`),
            apiFetch(`/users/${id}/comments?limit=20&offset=0&sort=desc`),
            apiFetch(`/users/${id}/post-likes?limit=20&offset=0&sort=desc`),
        ]);

        if (postsRes.ok) {
            const postsJson = await postsRes.json();
            const raw = Array.isArray(postsJson)
                ? postsJson
                : (postsJson.data ?? []);
            userPosts.value = mapPostsData(raw);
            postsOffset.value = userPosts.value.length;
            postsHasMore.value = userPosts.value.length >= 20;
        }
        if (repliesRes.ok) {
            const repliesJson = await repliesRes.json();
            const raw = Array.isArray(repliesJson)
                ? repliesJson
                : (repliesJson.data ?? []);
            userReplies.value = mapRepliesData(raw);
            repliesOffset.value = userReplies.value.length;
            repliesHasMore.value = userReplies.value.length >= 20;
        }
        if (likesRes.ok) {
            const likesJson = await likesRes.json();
            const raw = Array.isArray(likesJson)
                ? likesJson
                : (likesJson.data ?? []);
            userLikedPosts.value = mapLikesData(raw);
            likesOffset.value = userLikedPosts.value.length;
            likesHasMore.value = userLikedPosts.value.length >= 20;
        }
    } catch (error) {
        handleApiError(error, "Failed to load profile");
    } finally {
        loading.value = false;
        postsLoading.value = false;
        repliesLoading.value = false;
        likesLoading.value = false;
    }
}

function mapPostsData(raw: Record<string, unknown>[]): UserPost[] {
    return raw.map((p) => ({
        id: Number(p.post_id ?? p.id ?? 0),
        title: String(p.title ?? ""),
        tags: Array.isArray(p.tags) ? p.tags : [],
        comment_count: Number(p.comment_count ?? 0),
        like_count: Number(p.like_count ?? 0),
        view_count: Number(p.view_count ?? 0),
        created_at: String(p.created_at ?? ""),
    }));
}

function mapRepliesData(raw: Record<string, unknown>[]): UserReply[] {
    return raw.map((c) => ({
        id: Number(c.comment_id ?? c.id ?? 0),
        post_id: Number(c.post_id ?? 0),
        content: String(c.content ?? ""),
        created_at: String(c.created_at ?? ""),
    }));
}

function mapLikesData(raw: Record<string, unknown>[]): LikedPost[] {
    return raw.map((p) => ({
        post_id: Number(p.post_id ?? p.id ?? 0),
        author: String(p.author ?? ""),
        title: String(p.title ?? ""),
        tags: Array.isArray(p.tags) ? p.tags : [],
        comment_count: Number(p.comment_count ?? 0),
        like_count: Number(p.like_count ?? 0),
        created_at: String(p.created_at ?? ""),
    }));
}

async function apiGet(url: string): Promise<Record<string, unknown>[]> {
    const response = await apiFetch(url);
    if (!response.ok) throw new Error(`HTTP ${response.status}`);
    const json = await response.json();
    return Array.isArray(json) ? json : (json.data ?? []);
}

async function loadUserPosts() {
    if (postsLoading.value) return;
    postsLoading.value = true;
    postsHighlightId.value = null;
    try {
        const raw = await apiGet(
            `/users/${userId.value}/posts?limit=20&offset=0&sort=desc`,
        );
        userPosts.value = mapPostsData(raw);
        postsOffset.value = userPosts.value.length;
        postsHasMore.value = userPosts.value.length >= 20;
    } catch (error) {
        handleApiError(error, "Failed to load posts");
    } finally {
        postsLoading.value = false;
    }
}

async function loadMoreUserPosts() {
    if (postsLoadingMore.value || !postsHasMore.value) return;
    postsLoadingMore.value = true;
    try {
        const raw = await apiGet(
            `/users/${userId.value}/posts?limit=20&offset=${postsOffset.value}&sort=desc`,
        );
        const fetched = mapPostsData(raw);
        // Dedupe by id: a concurrent refresh can otherwise append the same
        // batch twice, duplicating rows and lighting up two highlights.
        const existingIds = new Set(userPosts.value.map((p) => p.id));
        const newPosts = fetched.filter((p) => !existingIds.has(p.id));
        if (newPosts.length > 0) {
            userPosts.value = [...userPosts.value, ...newPosts];
            postsOffset.value += newPosts.length;
            // Highlight the first item of the newly fetched batch only, not
            // the first item of the list as a whole.
            postsHighlightId.value = newPosts[0]!.id;
            setTimeout(() => {
                postsHighlightId.value = null;
            }, 2500);
        }
        if (fetched.length < 20) {
            postsHasMore.value = false;
        }
    } catch (error) {
        handleApiError(error, "Failed to load more posts");
    } finally {
        postsLoadingMore.value = false;
    }
}

async function loadUserReplies() {
    if (repliesLoading.value) return;
    repliesLoading.value = true;
    repliesHighlightId.value = null;
    try {
        const raw = await apiGet(
            `/users/${userId.value}/comments?limit=20&offset=0&sort=desc`,
        );
        userReplies.value = mapRepliesData(raw);
        repliesOffset.value = userReplies.value.length;
        repliesHasMore.value = userReplies.value.length >= 20;
    } catch (error) {
        handleApiError(error, "Failed to load replies");
    } finally {
        repliesLoading.value = false;
    }
}

async function loadMoreUserReplies() {
    if (repliesLoadingMore.value || !repliesHasMore.value) return;
    repliesLoadingMore.value = true;
    try {
        const raw = await apiGet(
            `/users/${userId.value}/comments?limit=20&offset=${repliesOffset.value}&sort=desc`,
        );
        const fetched = mapRepliesData(raw);
        const existingIds = new Set(userReplies.value.map((r) => r.id));
        const newReplies = fetched.filter((r) => !existingIds.has(r.id));
        if (newReplies.length > 0) {
            userReplies.value = [...userReplies.value, ...newReplies];
            repliesOffset.value += newReplies.length;
            repliesHighlightId.value = newReplies[0]!.id;
            setTimeout(() => {
                repliesHighlightId.value = null;
            }, 2500);
        }
        if (fetched.length < 20) {
            repliesHasMore.value = false;
        }
    } catch (error) {
        handleApiError(error, "Failed to load more replies");
    } finally {
        repliesLoadingMore.value = false;
    }
}

async function loadUserLikedPosts() {
    if (likesLoading.value) return;
    likesLoading.value = true;
    likesHighlightId.value = null;
    try {
        const raw = await apiGet(
            `/users/${userId.value}/post-likes?limit=20&offset=0&sort=desc`,
        );
        userLikedPosts.value = mapLikesData(raw);
        likesOffset.value = userLikedPosts.value.length;
        likesHasMore.value = userLikedPosts.value.length >= 20;
    } catch (error) {
        handleApiError(error, "Failed to load liked posts");
    } finally {
        likesLoading.value = false;
    }
}

async function loadMoreUserLikedPosts() {
    if (likesLoadingMore.value || !likesHasMore.value) return;
    likesLoadingMore.value = true;
    try {
        const raw = await apiGet(
            `/users/${userId.value}/post-likes?limit=20&offset=${likesOffset.value}&sort=desc`,
        );
        const fetched = mapLikesData(raw);
        const existingIds = new Set(userLikedPosts.value.map((p) => p.post_id));
        const newPosts = fetched.filter((p) => !existingIds.has(p.post_id));
        if (newPosts.length > 0) {
            userLikedPosts.value = [...userLikedPosts.value, ...newPosts];
            likesOffset.value += newPosts.length;
            // Highlight the first item of the newly fetched batch only.
            likesHighlightId.value = newPosts[0]!.post_id;
            setTimeout(() => {
                likesHighlightId.value = null;
            }, 2500);
        }
        if (fetched.length < 20) {
            likesHasMore.value = false;
        }
    } catch (error) {
        handleApiError(error, "Failed to load more liked posts");
    } finally {
        likesLoadingMore.value = false;
    }
}

// Only load more for the tab that is actually visible, so scrolling in the
// replies tab does not fire hidden background requests for the other tabs.
// All three tabs load more the same way Home (PostsList) does: fire as soon
// as the user scrolls within 200px of the bottom, with no debounce, and only
// from scroll events — never automatically on mount/activation. Only the
// visible tab loads, and only the first item of a load-more batch gets
// highlighted; the initial (first) request never highlights anything.
const postsLoadingMore = ref(false);
const repliesLoadingMore = ref(false);
const likesLoadingMore = ref(false);

function nearBottom() {
    const scrollHeight = document.documentElement.scrollHeight;
    const scrollTop = window.scrollY;
    const clientHeight = window.innerHeight;
    return scrollTop + clientHeight >= scrollHeight - 200;
}

function handleFeedScroll() {
    if (activeTab.value === "posts") {
        if (!postsLoadingMore.value && postsHasMore.value && nearBottom()) {
            loadMoreUserPosts();
        }
    } else if (activeTab.value === "likes") {
        if (!likesLoadingMore.value && likesHasMore.value && nearBottom()) {
            loadMoreUserLikedPosts();
        }
    } else if (activeTab.value === "replies") {
        if (!repliesLoadingMore.value && repliesHasMore.value && nearBottom()) {
            loadMoreUserReplies();
        }
    }
}

onActivated(() => {
    window.addEventListener("scroll", handleFeedScroll);
});

onDeactivated(() => {
    window.removeEventListener("scroll", handleFeedScroll);
});

onBeforeUnmount(() => {
    window.removeEventListener("scroll", handleFeedScroll);
});

function openPost(id: number) {
    router.push(`/posts/${id}`);
}

// Load on first mount, and only reload when the profile user actually changes.
// Returning to the same profile (e.g. back from a post) must NOT reload, otherwise
// the loading collapse breaks scroll restoration.
watch(
    () => route.params.userId,
    (id) => {
        const newId = Number(id);
        if (newId && newId !== loadedUserId) load();
    },
    { immediate: true },
);

// Insert fetched items that are not already in the list, keeping desc order
// (newest first). Returns the number of newly inserted items.
function mergeNewItems<T>(
    list: { value: T[] },
    offset: { value: number },
    fetched: T[],
    key: (item: T) => number,
): number {
    const existingIds = new Set(list.value.map(key));
    const fresh = fetched.filter((item) => !existingIds.has(key(item)));
    if (fresh.length === 0) return 0;
    list.value = [...fresh, ...list.value];
    offset.value += fresh.length;
    return fresh.length;
}

// Refresh data when re-activated from keep-alive (e.g. deleting a post on
// My Posts changes the counts shown here). Lists are NOT reset to page one:
// resetting collapses the page height and makes the restored scroll position
// jump (jitter). New items are merged in instead, so the user keeps their
// browsing context.
onActivated(async () => {
    const id = userId.value;
    if (!id || loading.value) return;

    // Clear any leftover load-more highlight from before the page was
    // deactivated, so returning never shows a stale highlight.
    postsHighlightId.value = null;
    repliesHighlightId.value = null;
    likesHighlightId.value = null;

    // Remember the scroll state so we can compensate if new items get
    // inserted at the top and push the content down.
    const prevScrollY = window.scrollY;
    const prevHeight = document.documentElement.scrollHeight;

    try {
        const [userRes, postsRes, repliesRes, likesRes] = await Promise.all([
            apiFetch(`/users/${id}`),
            apiFetch(`/users/${id}/posts?limit=20&offset=0&sort=desc`),
            apiFetch(`/users/${id}/comments?limit=20&offset=0&sort=desc`),
            apiFetch(`/users/${id}/post-likes?limit=20&offset=0&sort=desc`),
        ]);

        if (userRes.ok) {
            const json = await userRes.json();
            const data = json.data ?? json;
            user.value = {
                ...user.value,
                username: data.username ?? user.value.username,
                email: data.email ?? user.value.email,
                show_email: !!data.show_email,
                bio: data.bio ?? user.value.bio,
                links: Array.isArray(data.links)
                    ? data.links
                    : user.value.links,
                followers_count: Number(data.followers_count ?? 0),
                following_count: Number(data.following_count ?? 0),
                posts_count: Number(data.posts_count ?? 0),
                likes_count: Number(data.likes_count ?? 0),
                replies_count: Number(data.replies_count ?? 0),
            };
        }
        if (postsRes.ok) {
            const json = await postsRes.json();
            const raw = Array.isArray(json) ? json : (json.data ?? []);
            mergeNewItems(
                userPosts,
                postsOffset,
                mapPostsData(raw),
                (p) => p.id,
            );
        }
        if (repliesRes.ok) {
            const json = await repliesRes.json();
            const raw = Array.isArray(json) ? json : (json.data ?? []);
            mergeNewItems(
                userReplies,
                repliesOffset,
                mapRepliesData(raw),
                (r) => r.id,
            );
        }
        if (likesRes.ok) {
            const json = await likesRes.json();
            const raw = Array.isArray(json) ? json : (json.data ?? []);
            mergeNewItems(
                userLikedPosts,
                likesOffset,
                mapLikesData(raw),
                (p) => p.post_id,
            );
        }
    } catch {
        // Silently ignore errors on activation refresh
    }

    // New items inserted at the top push content down; compensate so the
    // viewport keeps showing the same rows instead of jumping.
    await nextTick();
    const newHeight = document.documentElement.scrollHeight;
    if (newHeight > prevHeight) {
        window.scrollTo(0, prevScrollY + (newHeight - prevHeight));
    }
});
</script>

<style scoped>
.user-profile-page {
    min-height: 100vh;
    padding: 32px 0 80px;
}

.profile-card {
    width: 75%;
    margin: 0 auto;
    padding: 48px 24px 8px;
    display: flex;
    flex-direction: column;
    align-items: center;
    text-align: center;
}

.profile-header {
    width: 100%;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 6px;
}

.avatar-wrapper {
    position: relative;
    display: inline-block;
}

.avatar-wrapper.own {
    cursor: pointer;
}

.avatar-overlay {
    position: absolute;
    inset: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 50%;
    background: rgba(0, 0, 0, 0.5);
    opacity: 0;
    transition: opacity 0.2s ease;
    cursor: pointer;
}

.avatar-wrapper.own:hover .avatar-overlay {
    opacity: 1;
}

.hidden-input {
    display: none;
}

.camera-icon {
    width: 28px;
    height: 28px;
    color: #ffffff;
}

.avatar-delete-btn {
    position: absolute;
    left: -4px;
    bottom: -4px;
    width: 24px;
    height: 24px;
    border-radius: 50%;
    border: 2px solid #1a1a1a;
    background: #da3633;
    color: #ffffff;
    font-size: 16px;
    font-weight: 700;
    line-height: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    opacity: 0;
    transition: opacity 0.2s ease;
    padding: 0;
}

.avatar-wrapper.own:hover .avatar-delete-btn {
    opacity: 1;
}

.avatar-delete-btn:hover {
    background: #f85149;
}

.avatar-delete-btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
}

.username {
    margin: 18px 0 0;
    font-size: 28px;
    font-weight: 700;
    line-height: 1.2;
    color: #e4e6e8;
}

.handle {
    font-size: 15px;
    color: #8c8c8c;
}

.email {
    margin: 0;
    font-size: 13px;
    line-height: 1.5;
    color: #8c8c8c;
}

.bio {
    margin: 8px 0 0;
    max-width: 560px;
    font-size: 14px;
    line-height: 1.6;
    color: #e4e6e8;
    word-break: break-word;
}

.meta-row {
    display: flex;
    align-items: center;
    justify-content: center;
    flex-wrap: wrap;
    gap: 6px 22px;
    margin-top: 10px;
    font-size: 13px;
    color: #8c8c8c;
}

.meta-item {
    display: inline-flex;
    align-items: center;
    gap: 6px;
}

.info-icon {
    width: 14px;
    height: 14px;
    flex-shrink: 0;
    color: #8c8c8c;
}

.website {
    color: #6cbbf7;
    text-decoration: none;
    word-break: break-all;
}

.website:hover {
    text-decoration: underline;
}

.not-found-title {
    margin: 0;
    font-size: 22px;
    font-weight: 700;
    color: #e4e6e8;
}

.not-found-hint {
    margin: 0;
    font-size: 13px;
    color: #8c8c8c;
}

.stats-row {
    display: flex;
    align-items: center;
    justify-content: center;
    flex-wrap: wrap;
    gap: 8px 36px;
    width: 100%;
    max-width: 660px;
    margin-top: 20px;
    padding: 18px 0 2px;
    border-top: 1px solid #262626;
}

.stat {
    display: inline-flex;
    flex-direction: column;
    align-items: center;
    gap: 3px;
    min-width: 62px;
}

.stat strong {
    font-size: 20px;
    font-weight: 700;
    color: #e4e6e8;
    font-variant-numeric: tabular-nums;
}

.stat-label {
    font-size: 12px;
    color: #8c8c8c;
}

.stat.clickable {
    cursor: pointer;
}

.stat.clickable:hover strong {
    color: #6cbbf7;
}

.profile-tabs {
    width: 75%;
    margin: 24px auto 0;
}

.tabs-bar {
    display: flex;
    justify-content: center;
    gap: 0;
    border-bottom: 1px solid #262626;
}

.tab-btn {
    padding: 12px 28px;
    background: transparent;
    border: none;
    border-bottom: 2px solid transparent;
    color: #8c8c8c;
    font-size: 14px;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.15s ease;
}

.tab-btn:hover {
    color: #e4e6e8;
}

.tab-btn.active {
    color: #e4e6e8;
    border-bottom-color: #e4e6e8;
}

.tabs-content {
    margin-top: 8px;
}

.tab-panel {
    min-height: 120px;
}

.feed {
    display: flex;
    flex-direction: column;
}

.topic-row {
    padding: 14px 0;
    border-bottom: 1px solid #262626;
    cursor: pointer;
    transition: background 0.15s ease;
}

.topic-row:first-child {
    border-top: 1px solid #262626;
}

.topic-row:hover {
    background: rgba(255, 255, 255, 0.03);
}

.topic-top {
    display: flex;
    align-items: center;
    gap: 12px;
}

.topic-title {
    flex: 1;
    min-width: 0;
    margin: 0;
    font-size: 16px;
    font-weight: 600;
    line-height: 1.4;
    color: #e4e6e8;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.topic-stats {
    flex-shrink: 0;
    display: flex;
    align-items: center;
    gap: 14px;
}

.topic-stat {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    font-size: 13px;
    color: #8c8c8c;
    white-space: nowrap;
}

.stat-icon {
    width: 15px;
    height: 15px;
    fill: none;
    stroke: currentColor;
    stroke-width: 2;
}

.like-icon {
    fill: none;
    stroke: #8c8c8c;
    stroke-width: 2;
}

.topic-bottom {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    gap: 6px;
    margin-top: 6px;
    padding-left: 2px;
    font-size: 13px;
    color: #8c8c8c;
}

.topic-author {
    color: #e4e6e8;
    font-weight: 500;
}

.topic-meta {
    display: flex;
    align-items: center;
    gap: 6px;
    margin-top: 6px;
    font-size: 13px;
    color: #8c8c8c;
}

.topic-date {
    color: #8c8c8c;
}

.meta-dot {
    color: #595959;
}

.topic-tag {
    padding: 1px 8px;
    background: #1f1f1f;
    border: 1px solid #333;
    border-radius: 4px;
    font-size: 12px;
    color: #8c8c8c;
    white-space: nowrap;
}

.reply-row {
    display: block;
}

.reply-context {
    display: flex;
    align-items: center;
    gap: 6px;
    font-size: 13px;
    color: #8c8c8c;
}

.reply-icon {
    width: 16px;
    height: 16px;
    fill: currentColor;
    flex-shrink: 0;
}

.reply-hint {
    color: #8c8c8c;
}

.reply-content {
    margin: 8px 0 0;
    font-size: 14px;
    line-height: 1.5;
    color: #e4e6e8;
    overflow: hidden;
    text-overflow: ellipsis;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
}

.empty-hint {
    text-align: center;
    padding: 32px 16px;
    color: #8c8c8c;
    font-size: 14px;
}

.topic-row.new-item,
.reply-row.new-item {
    animation: highlight-flash 2.5s ease-out;
}

@keyframes highlight-flash {
    0% {
        background-color: rgba(64, 158, 255, 0.15);
    }
    100% {
        background-color: transparent;
    }
}

.loading-more {
    text-align: center;
    padding: 16px;
    color: #8c8c8c;
    font-size: 14px;
}
</style>
