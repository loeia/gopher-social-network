<template>
    <div class="feed" v-loading="loading">
        <div
            v-for="post in posts"
            :key="post.post_id"
            class="card"
            @click="openPost(post.post_id)"
        >
            <div class="card-header">
                <h2 class="card-title">{{ post.title }}</h2>
                <div class="card-meta">
                    <svg
                        class="stat-icon like-icon"
                        viewBox="0 0 24 24"
                        aria-hidden="true"
                    >
                        <path
                            d="M12 21.35l-1.45-1.32C5.4 15.36 2 12.28 2 8.5 2 5.42 4.42 3 7.5 3c1.74 0 3.41.81 4.5 2.09C13.09 3.81 14.76 3 16.5 3 19.58 3 22 5.42 22 8.5c0 3.78-3.4 6.86-8.55 11.54L12 21.35z"
                        />
                    </svg>
                    <span class="count">{{ post.like_count ?? 0 }}</span>
                    <svg
                        class="stat-icon comment-icon"
                        viewBox="0 0 24 24"
                        aria-hidden="true"
                    >
                        <path
                            d="M20 2H4a2 2 0 0 0-2 2v18l4-4h14a2 2 0 0 0 2-2V4a2 2 0 0 0-2-2z"
                        />
                    </svg>
                    <span class="count">{{ post.comment_count ?? 0 }}</span>
                    <span class="card-date">{{
                        formatDate(post.created_at)
                    }}</span>
                </div>
            </div>
            <div v-if="post.tags && post.tags.length" class="card-tags">
                <span v-for="tag in post.tags" :key="tag" class="tag-pill">{{
                    tag
                }}</span>
            </div>
            <div class="card-author">
                <UserAvatar
                    :user-id="post.user_id"
                    :username="post.author"
                    :size="28"
                />
                <span>{{ post.author }}</span>
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
import {
    nextTick,
    onActivated,
    onBeforeUnmount,
    onDeactivated,
    onMounted,
    ref,
} from "vue";
import { useRouter } from "vue-router";
import { storeToRefs } from "pinia";
import { useFeedStore } from "@/stores/feed";
import { notify } from "@/utils/message";
import UserAvatar from "@/components/UserAvatar.vue";

const store = useFeedStore();
const { likedPosts: posts } = storeToRefs(store);
const loading = ref(false);

const router = useRouter();

function openPost(id: number) {
    router.push(`/posts/${id}`);
}

function formatDate(value?: string) {
    if (!value) return "";
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) return value;
    return date.toLocaleString();
}

function saveScroll() {
    store.likedScrollTop = window.scrollY;
}

function restoreScroll() {
    nextTick(() => window.scrollTo({ top: store.likedScrollTop }));
}

async function loadLikedPosts() {
    loading.value = true;
    try {
        await store.fetchLikedPosts();
    } catch (error) {
        console.error("Load liked posts error:", error);
        notify("error", "Failed to load liked posts");
    } finally {
        loading.value = false;
    }
}

onMounted(() => {
    restoreScroll();
    loadLikedPosts();
});
onActivated(() => {
    restoreScroll();
    loadLikedPosts();
});
onDeactivated(saveScroll);
onBeforeUnmount(saveScroll);
</script>

<style scoped>
.feed {
    margin: 0 20%;
    display: flex;
    flex-direction: column;
    gap: 20px;
}

.card {
    background: #141414;
    border: 1px solid #262626;
    border-radius: 12px;
    padding: 24px;
    cursor: pointer;
    transition:
        border-color 0.2s ease,
        transform 0.2s ease;
}

.card:hover {
    border-color: #ffffff;
    transform: translateY(-2px);
}

.card-header {
    display: flex;
    align-items: baseline;
    justify-content: space-between;
    gap: 16px;
}

.card-meta {
    flex-shrink: 0;
    display: flex;
    align-items: center;
    gap: 5px;
    font-size: 13px;
    color: #8c8c8c;
}

.stat-icon {
    width: 16px;
    height: 16px;
}

.like-icon path {
    fill: #e05c5c;
    stroke: #e05c5c;
}

.comment-icon path {
    fill: none;
    stroke: currentColor;
    stroke-width: 2;
}

.card-meta .card-date {
    margin-left: 8px;
}

.card-title {
    flex: 1;
    min-width: 0;
    margin: 0;
    font-size: 20px;
    font-weight: 600;
    color: #ffffff;
    overflow-wrap: break-word;
    word-break: break-word;
}

.card-date {
    flex-shrink: 0;
    font-size: 13px;
    color: #8c8c8c;
}

.card-tags {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    padding-top: 12px;
}

.tag-pill {
    padding: 2px 10px;
    border: 1px solid #3d444d;
    border-radius: 999px;
    font-size: 12px;
    color: #bfbfbf;
    background: #1f1f1f;
    white-space: nowrap;
}

.card-author {
    display: flex;
    align-items: center;
    gap: 10px;
    padding-top: 16px;
    border-top: 1px solid #262626;
    font-size: 14px;
    color: #8c8c8c;
}
</style>
