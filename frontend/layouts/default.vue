<script setup lang="ts">
const auth = useAuthStore();
// Make sure user is loaded on first paint after login redirect
onMounted(async () => {
  if (auth.token && !auth.user) await auth.fetchMe();
});
</script>

<template>
  <div class="min-h-screen flex p-3.5 gap-3.5">
    <Sidebar />
    <main class="flex-1 lg:ml-[274px] flex flex-col gap-3.5 min-w-0">
      <Topbar />
      <div class="flex-1">
        <slot />
      </div>
    </main>
  </div>
</template>
