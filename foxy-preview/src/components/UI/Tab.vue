<script setup lang="ts">
defineProps<{
	modelValue: string,
	value: string,
}>();

const emit = defineEmits<{
	(e: 'update:modelValue', value: string): void;
}>();
</script>
<template>
	<a class="cursor-pointer px-2 py-1 uppercase font-bold text-xxs" @click="emit('update:modelValue', value)" :class="{
						'bg-neutral-100 text-neutral-600': modelValue === value,
						'bg-neutral-300 text-neutral-500': modelValue !== value,
					}"><slot :selected="modelValue === value"></slot></a>
</template>