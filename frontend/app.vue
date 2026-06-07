<script setup lang="ts">
useHead({
  titleTemplate: (t) => (t ? `${t} | Somon CRM` : "Somon CRM"),
});
</script>

<template>
  <NuxtLoadingIndicator color="#1f7a4d" :height="3" />
  <NuxtLayout>
    <NuxtPage />
  </NuxtLayout>
  <Toaster />
</template>
