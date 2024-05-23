<script setup lang="ts">
import { computed } from 'vue';

const props = defineProps<{
	title: string,
	modelValue: boolean,
}>();

const emit = defineEmits<{
	(e: 'update:modelValue', value: boolean): void;
}>();

const currentValue = computed({
	get: () => props.modelValue,
	set: (value) => emit('update:modelValue', value),
});
</script>
<template>
	<label class="inline-flex items-center cursor-pointer">
		<input type="checkbox" v-model="currentValue" class="sr-only peer">
		<div class="relative w-9 h-5 bg-gray-200 peer-focus:outline-none peer-focus:ring-2 peer-focus:ring-blue-300 dark:peer-focus:ring-blue-800 rounded-full peer dark:bg-gray-700 peer-checked:after:translate-x-full rtl:peer-checked:after:-translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:start-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-4 after:w-4 after:transition-all dark:border-gray-600 peer-checked:bg-blue-600"></div>
		<span class="ms-2 text-xs font-medium text-gray-900 dark:text-gray-300">{{ title }}</span>
	</label>
</template>