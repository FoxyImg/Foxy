<script setup lang="ts">
import {computed} from "vue";
import SmallLabel from "@/components/UI/SmallLabel.vue";

import pDebounce from 'p-debounce';

const props =defineProps<{
	label: string,
	modelValue: string|null,
}>();

const emit = defineEmits<{
	(e: 'update:modelValue', value: string|null): void;
}>();

function updateModelValue(value: string|null) {
	emit('update:modelValue', value);
}

const debouncedUpdateModelValue = pDebounce(updateModelValue, 500);

const currentValue = computed({
	get: () => props.modelValue,
	set: (value) => debouncedUpdateModelValue(value),
});
</script>
<template>
	<div class="flex flex-col gap-1">
		<SmallLabel>{{ label }}</SmallLabel>
		<input type="text" class="w-full border border-neutral-200 rounded-md px-2 py-1 text-sm" v-model="currentValue" />
	</div>
</template>