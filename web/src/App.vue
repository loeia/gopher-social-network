<template>
  <div class="layout">
    <NavBar />
    <keep-alive include="HomePage">
      <router-view />
    </keep-alive>
  </div>
</template>

<script setup lang="ts">
import NavBar from '@/components/NavBar.vue'
</script>

<style>
body {
  margin: 0;
  background: #0a0a0a;
  color: #ffffff;
}
</style>
