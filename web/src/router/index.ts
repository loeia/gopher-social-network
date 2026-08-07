import { createRouter, createWebHistory } from 'vue-router'
import HomePage from '@/views/HomePage.vue'
import ConfirmationPage from '@/views/ConfirmationPage.vue'
import PostDetailPage from '@/views/PostDetailPage.vue'
import CreatePostPage from '@/views/CreatePostPage.vue'
import LoginPage from '@/views/LoginPage.vue'

const router = createRouter({
history: createWebHistory(import.meta.env.BASE_URL),
  scrollBehavior(to, from, savedPosition) {
    // Restore position on back/forward, otherwise always scroll to top.
    if (savedPosition) return savedPosition
    if (to.name === 'Home') return false
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
      path: '/login',
      name: 'Login',
      component: LoginPage,
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
    },
    ]
})

export default router
