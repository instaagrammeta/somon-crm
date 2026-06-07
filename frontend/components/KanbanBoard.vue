<script setup lang="ts">
/** Generic kanban renderer for both requests and lids.
 *  Drag-and-drop is plain HTML5 (no extra deps).
 */
interface Column {
  id: number;
  title: string;
  color?: string;
}
interface Card {
  id: number;
  column_id: number;
  order_index: number;
}

const props = defineProps<{
  columns: Column[];
  cards: Card[];
  loading?: boolean;
  /** Card render scope slot key */
}>();

const emit = defineEmits<{
  (e: "moveCard", payload: { cardId: number; toColumnId: number; orderIndex: number }): void;
  (e: "addCard", columnId: number): void;
  (e: "editColumn", col: Column): void;
  (e: "deleteColumn", col: Column): void;
  (e: "addColumn"): void;
}>();

const cardsByColumn = computed(() => {
  const m = new Map<number, Card[]>();
  for (const c of props.cards) {
    if (!m.has(c.column_id)) m.set(c.column_id, []);
    m.get(c.column_id)!.push(c);
  }
  for (const list of m.values()) {
    list.sort((a, b) => (a.order_index ?? 0) - (b.order_index ?? 0));
  }
  return m;
});

const dragId = ref<number | null>(null);
const overCol = ref<number | null>(null);

const onDragStart = (e: DragEvent, id: number) => {
  dragId.value = id;
  e.dataTransfer?.setData("text/plain", String(id));
};
const onDragOverCol = (e: DragEvent, id: number) => {
  e.preventDefault();
  overCol.value = id;
};
const onDrop = (e: DragEvent, columnId: number) => {
  e.preventDefault();
  const id = dragId.value || Number(e.dataTransfer?.getData("text/plain") || 0);
  if (!id) return;
  const target = cardsByColumn.value.get(columnId) || [];
  emit("moveCard", { cardId: id, toColumnId: columnId, orderIndex: target.length });
  dragId.value = null;
  overCol.value = null;
};
</script>

<template>
  <div class="flex gap-4 overflow-x-auto pb-2 min-h-[500px]">
    <div
      v-for="col in columns"
      :key="col.id"
      class="bg-page rounded-2xl p-3 w-[300px] shrink-0 flex flex-col"
      :class="{ 'ring-2 ring-brand': overCol === col.id }"
      @dragover="onDragOverCol($event, col.id)"
      @drop="onDrop($event, col.id)"
    >
      <header class="flex items-center justify-between gap-2 mb-3 px-1">
        <div class="flex items-center gap-2 flex-1 min-w-0">
          <span class="w-2 h-2 rounded-full shrink-0" :style="{ background: col.color || '#3b82f6' }" />
          <span class="font-semibold text-ink text-sm truncate">{{ col.title }}</span>
          <span class="text-xs text-ink-soft">
            {{ (cardsByColumn.get(col.id) || []).length }}
          </span>
        </div>
        <div class="flex items-center gap-0.5">
          <button class="btn-ghost !p-1.5 text-xs" @click="emit('editColumn', col)">
            <i class="fa-solid fa-pen" />
          </button>
          <button class="btn-ghost !p-1.5 text-xs text-red-500" @click="emit('deleteColumn', col)">
            <i class="fa-solid fa-trash" />
          </button>
        </div>
      </header>

      <div class="flex flex-col gap-2 flex-1 min-h-[100px]">
        <div
          v-for="card in cardsByColumn.get(col.id) || []"
          :key="card.id"
          class="bg-white rounded-xl p-3 shadow-card border border-border-soft cursor-grab active:cursor-grabbing"
          draggable="true"
          @dragstart="onDragStart($event, card.id)"
        >
          <slot name="card" :card="card" :column="col" />
        </div>
      </div>

      <button class="btn-ghost !justify-start mt-2 text-xs" @click="emit('addCard', col.id)">
        <i class="fa-solid fa-plus text-[10px]" /> {{ $t("app.create") }}
      </button>
    </div>

    <button
      class="bg-white border-2 border-dashed border-border-soft rounded-2xl w-[300px] shrink-0 hover:border-brand hover:text-brand text-ink-soft font-semibold transition-colors"
      @click="emit('addColumn')"
    >
      <i class="fa-solid fa-plus mr-2" />
      {{ $t("kanban.create_column") }}
    </button>
  </div>
</template>
