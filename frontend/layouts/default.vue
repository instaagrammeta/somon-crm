<script setup lang="ts">
const auth = useAuthStore();
onMounted(async () => {
  if (auth.token && !auth.user) await auth.fetchMe();
});
</script>

<template>
  <!-- app-container: 14px padding + gap, like base.html -->
  <div class="min-h-screen flex p-3.5 gap-3.5">
    <Sidebar />

    <!-- main-area: offset by sidebar width on desktop -->
    <div class="flex-1 lg:ml-[274px] flex flex-col gap-3.5 min-w-0">
      <Topbar />
      <!-- main-content card -->
      <main class="bg-white rounded-3xl p-6 lg:px-7 lg:py-6 shadow-card flex-1" style="min-height: calc(100vh - 130px)">
        <div class="max-w-[1400px] mx-auto">
          <slot />
        </div>
      </main>
    </div>
  </div>
</template>
