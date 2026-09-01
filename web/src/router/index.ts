import { createRouter, createWebHistory } from 'vue-router'
import { clearToken, getToken, isTokenExpired } from '@/api'
import { notify } from '@/utils/message'
import HomePage from '@/views/HomePage.vue'
import ConfirmationPage from '@/views/ConfirmationPage.vue'
import PostDetailPage from '@/views/PostDetailPage.vue'
import CreatePostPage from '@/views/CreatePostPage.vue'
import MyPostsPage from '@/views/MyPostsPage.vue'
import PostLikesPage from '@/views/PostLikesPage.vue'
import CommentLikesPage from '@/views/CommentLikesPage.vue'
import FollowingPage from '@/views/FollowingPage.vue'
import FollowersPage from '@/views/FollowersPage.vue'
import UserProfilePage from '@/views/UserProfilePage.vue'
import LoginPage from '@/views/LoginPage.vue'
import SignUpPage from '@/views/SignUpPage.vue'
import SearchResults from '@/views/SearchResults.vue'
import NotFoundPage from '@/views/NotFoundPage.vue'
import ForgetPasswordPage from '@/views/ForgetPasswordPage.vue'
import ResetPasswordPage from '@/views/ResetPasswordPage.vue'
import SettingsPage from '@/views/SettingsPage.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  scrollBehavior(to, _from, savedPosition) {
    if (savedPosition) return { ...savedPosition, behavior: 'instant' }
    if (to.hash) return false
    return { top: 0 }
  },
  routes: [
    {
      path: '/',
      name: 'Home',
      component: HomePage,
    },
    {
      path: '/posts/new',
      name: 'CreatePost',
      component: CreatePostPage,
    },
    {
      path: '/my-posts',
      name: 'MyPosts',
      component: MyPostsPage,
    },
    {
      path: '/post-likes',
      name: 'PostLikes',
      component: PostLikesPage,
    },
    {
      path: '/comment-likes',
      name: 'CommentLikes',
      component: CommentLikesPage,
    },
    {
      path: '/users/:userId/following',
      name: 'Following',
      component: FollowingPage,
    },
    {
      path: '/users/:userId/followers',
      name: 'Followers',
      component: FollowersPage,
    },
    {
      path: '/users/:userId',
      name: 'UserProfile',
      component: UserProfilePage,
    },
    {
      path: '/login',
      name: 'Login',
      component: LoginPage,
      meta: { hideNavBar: true },
    },
    {
      path: '/search',
      name: 'SearchResults',
      component: SearchResults,
    },
    {
      path: '/signup',
      name: 'SignUp',
      component: SignUpPage,
      meta: { hideNavBar: true },
    },
    {
      path: '/forgot-password',
      name: 'ForgotPassword',
      component: ForgetPasswordPage,
      meta: { hideNavBar: true },
    },
    {
      path: '/reset-password/:token',
      name: 'ResetPassword',
      component: ResetPasswordPage,
      meta: { hideNavBar: true },
    },
    {
      path: '/settings',
      name: 'Settings',
      component: SettingsPage,
    },
    {
      path: '/posts/:postId',
      name: 'PostDetail',
      component: PostDetailPage,
    },
    {
      path: '/confirm/:token',
      name: 'Confirmation',
      component: ConfirmationPage,
      meta: { hideNavBar: true },
    },
    {
      path: '/:pathMatch(.*)*',
      name: 'NotFound',
      component: NotFoundPage,
      meta: { hideNavBar: true },
    },
  ],
})

router.beforeEach((to) => {
  const currentToken = getToken()
  if (currentToken && isTokenExpired(currentToken)) {
    clearToken()
    if (to.name !== 'Login') {
      notify('warning', 'Session expired, please sign in again')
      return { name: 'Login' }
    }
  }
  return true
})

export default router
