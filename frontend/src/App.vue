<script setup lang="ts">
import { RouterLink, RouterView } from 'vue-router'
</script>

<template>
  <header>
    <div class="wrapper">
      <h1 class="title">License管理系统</h1>
    </div>
  </header>

  <RouterView />
</template>

<style scoped>
header {
  line-height: 1.2;
  background-color: #f5f5f5;
  padding: 0.5rem 0;
  box-shadow: 0 2px 4px rgba(0,0,0,0.1);
}

.title {
  margin: 0 0 0.5rem 0;
  text-align: center;
  color: #409eff;
  font-size: 1.5rem;
}

.wrapper {
  width: 100%;
  margin: 0;
  padding: 0 1rem;
}

nav {
  width: 100%;
  font-size: 16px;
  text-align: center;
  margin-top: 0.5rem;
}

nav a.router-link-exact-active {
  color: var(--color-text);
  font-weight: bold;
}

nav a.router-link-exact-active:hover {
  background-color: transparent;
}

nav a {
  display: inline-block;
  padding: 0 1rem;
  border-left: 1px solid var(--color-border);
  color: #409eff;
  text-decoration: none;
}

nav a:first-of-type {
  border: 0;
}

nav a:hover {
  color: #337ecc;
}

@media (min-width: 1024px) {
  header {
    display: flex;
    place-items: center;
    padding-right: calc(var(--section-gap) / 2);
    height: auto;
  }

  header .wrapper {
    display: flex;
    place-items: flex-start;
    flex-wrap: wrap;
    width: 100%;
  }

  .title {
    margin: 0 1rem 0 0;
    font-size: 1.5rem;
  }

  nav {
    text-align: left;
    margin-left: -1rem;
    font-size: 1rem;
    padding: 0;
    margin-top: 0;
  }
}
</style>
