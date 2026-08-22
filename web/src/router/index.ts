import { createRouter, createWebHistory } from "vue-router";
import { clearToken, getToken, isTokenExpired } from "@/api";
import { notify } from "@/utils/message";
import HomePage from "@/views/HomePage.vue";
import ConfirmationPage from "@/views/ConfirmationPage.vue";
import PostDetailPage from "@/views/PostDetailPage.vue";
import CreatePostPage from "@/views/CreatePostPage.vue";
import MyPostsPage from "@/views/MyPostsPage.vue";
import UserProfilePage from "@/views/UserProfilePage.vue";
import LoginPage from "@/views/LoginPage.vue";
import SignUpPage from "@/views/SignUpPage.vue";
import SearchResults from "@/views/SearchResults.vue";
import NotFoundPage from "@/views/NotFoundPage.vue";
import ForgetPasswordPage from "@/views/ForgetPasswordPage.vue";
import ResetPasswordPage from "@/views/ResetPasswordPage.vue";

const router = createRouter({
    history: createWebHistory(import.meta.env.BASE_URL),
    scrollBehavior(to, from, savedPosition) {
        // Restore position on back/forward. Keep-alive feed pages manage their own
        // scroll (via per-view scroll state). Everything else starts at the top.
        if (savedPosition) return savedPosition;
        if (
            to.name === "Home" ||
            to.name === "MyPosts" ||
            to.name === "SearchResults"
        )
            return false;
        return { top: 0 };
    },
    routes: [
        {
            path: "/",
            name: "Home",
            component: HomePage,
        },
        {
            path: "/posts/new",
            name: "CreatePost",
            component: CreatePostPage,
        },
        {
            path: "/my-posts",
            name: "MyPosts",
            component: MyPostsPage,
        },
        {
            path: "/users/:userId",
            name: "UserProfile",
            component: UserProfilePage,
        },
        {
            path: "/login",
            name: "Login",
            component: LoginPage,
        },
        {
            path: "/search",
            name: "SearchResults",
            component: SearchResults,
        },
        {
            path: "/signup",
            name: "SignUp",
            component: SignUpPage,
        },
        {
            path: "/forgot-password",
            name: "ForgotPassword",
            component: ForgetPasswordPage,
        },
        {
            path: "/reset-password/:token",
            name: "ResetPassword",
            component: ResetPasswordPage,
            meta: { hideNavBar: true },
        },
        {
            path: "/posts/:postId",
            name: "PostDetail",
            component: PostDetailPage,
        },
        {
            path: "/confirm/:token",
            name: "Confirmation",
            component: ConfirmationPage,
            meta: { hideNavBar: true },
        },
        {
            path: "/:pathMatch(.*)*",
            name: "NotFound",
            component: NotFoundPage,
            meta: { hideNavBar: true },
        },
    ],
});

router.beforeEach((to) => {
    const currentToken = getToken();
    if (currentToken && isTokenExpired(currentToken)) {
        clearToken();
        if (to.name !== "Login") {
            notify("warning", "Session expired, please sign in again");
            return { name: "Login" };
        }
    }
    return true;
});

export default router;
