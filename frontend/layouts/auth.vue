<template>
  <div class="min-h-screen bg-page flex items-center justify-center p-0 sm:p-5 relative overflow-hidden">
    <!-- Decorative gradient blobs (match login.html body::before/after) -->
    <span class="fixed -top-[200px] -right-[200px] w-[600px] h-[600px] rounded-full bg-gradient-to-br from-brand-light to-transparent z-0" />
    <span class="fixed -bottom-[200px] -left-[200px] w-[500px] h-[500px] rounded-full bg-gradient-to-br from-brand-soft to-transparent z-0" />
    <slot />
  </div>
</template>
