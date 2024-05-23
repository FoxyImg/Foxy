<script setup lang="ts">
import { computed } from 'vue';

const props = defineProps<{
	title: string,
	modelValue: string|null,
	default: string|null,
	options: { [key:string]: string }
}>();

const emit = defineEmits<{
	(e: 'update:modelValue', value: string|null): void;
}>();

const currentValue = computed({
	get: () => props.modelValue,
	set: (value) => emit('update:modelValue', value),
});
</script>
<template>
	<div class="flex flex-col gap-1.5">
		<div class="flex flex-col items-start gap-1">
			<label class="text-xxs uppercase text-neutral-600" :class="{'font-bold text-neutral-700': currentValue !== props.default}">{{title}}</label>
			<select v-model="currentValue" class="w-full border border-neutral-200 text-xs flex-1 rounded-md py-1.5 px-1">
				<option :selected="currentValue === null" :value="null">None</option>
				<option v-for="(param, key, index) in options" :key="key" :value="key" :selected="key === currentValue">{{ param }}</option>
			</select>
		</div>
		<div class="flex justify-end text-xxxs uppercase">
			<a href="#" @click.prevent.stop="emit('update:modelValue', props.default)">Reset</a>
		</div>
	</div>
</template>