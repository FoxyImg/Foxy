<script setup lang="ts">
import { computed, ref, watch } from 'vue';

const props = defineProps<{
	modelValue: number|null,
	title: string;
	min: number;
	max: number;
	step: number;
	defaultLabel: string|undefined;
	default: number|null;
	suffix?: string;
}>();

const emit = defineEmits<{
	(e: 'update:modelValue', value: number|null): void;
}>();

const currentValue = computed({
	get: () => props.modelValue,
	set: (value) => emit('update:modelValue', value),
});

const editing = ref(false);
const inputRef = ref<HTMLInputElement|null>(null);

watch(inputRef, (el) => {
	if (el) {
		el.focus();
		el.select();
	}
});

</script>
<template>
	<div class="flex flex-col gap-1.5">
		<div class="flex items-center justify-between text-xs text-neutral-600">
			<label :class="{'font-bold text-neutral-700': currentValue !== props.default}">{{  title  }}</label>
			<input
				v-if="editing"
				ref="inputRef"
				type="number"
				v-model.number="currentValue"
				:min="min"
				:max="max"
				:step="step"
				@focusout="editing=false"
				@keyup.enter="editing=false" autofocus
			/>
			<div v-else @click="editing = true" class="cursor-pointer underline decoration-dotted text-xxs" :class="{'font-bold text-neutral-700': currentValue !== props.default}">
				<span v-if="defaultLabel && currentValue === props.default">{{ defaultLabel }}</span>
				<span v-else>{{ currentValue }}{{ props.suffix }}</span>
			</div>
		</div>
		<input
			class="range"
			:class="{'in-use': currentValue !== props.default }"
			type="range"
			:min="min"
			:max="max"
			:step="step"
			v-model.number="currentValue"
		/>
		<div class="flex justify-end text-xxxs font-bold uppercase">
			<a href="#" @click.prevent.stop="emit('update:modelValue', props.default)">Reset</a>
		</div>
	</div>
</template>
<style>
.range {
	appearance: none;
}

.range:focus {
	outline: none;
}

.range:focus::-webkit-slider-runnable-track, .range.in-use::-webkit-slider-runnable-track {
	background: theme('colors.neutral.300');
}

.range:focus::-moz-range-track {
	background: theme('colors.neutral.300');
}


.range::-webkit-slider-runnable-track {
	appearance: none;
	width: 100%;
	height: 15px;
	cursor: pointer;
	background: theme('colors.neutral.200');
	border-radius: 15px;
	box-shadow: none;
}

.range::-moz-range-track {
	appearance: none;
	width: 100%;
	height: 15px;
	cursor: pointer;
	background: theme('colors.neutral.200');
	border-radius: 15px;
	box-shadow: none;
}

.range::-webkit-slider-thumb {
	appearance: none;
	cursor: pointer;
	background: theme('colors.neutral.400');
	border: 1px solid theme('colors.neutral.500');
	width: 15px;
	height: 15px;
	border-radius: 15px;
	margin-top: 0px;
}

.range::-moz-range-thumb {
	appearance: none;
	cursor: pointer;
	background: theme('colors.neutral.400');
	border: 1px solid theme('colors.neutral.500');
	width: 15px;
	height: 15px;
	border-radius: 15px;
	margin-top: 0px;
}

.range:focus::-webkit-slider-thumb, .range.in-use::-webkit-slider-thumb {
	appearance: none;
	cursor: pointer;
	background: #2271b1;
	border: 1px solid #2271b1;
	width: 15px;
	height: 15px;
	border-radius: 15px;
	margin-top: 0px;
}

.range:focus::-moz-slider-thumb, .range.in-use::-moz-slider-thumb {
	appearance: none;
	cursor: pointer;
	background: #2271b1;
	border: 1px solid #2271b1;
	width: 15px;
	height: 15px;
	border-radius: 15px;
	margin-top: 0px;
}
</style>