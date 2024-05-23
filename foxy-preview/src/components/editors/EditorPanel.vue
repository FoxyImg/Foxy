<script setup lang="ts">
import SmallLabel from "@/components/UI/SmallLabel.vue";

defineProps< {
	title: string
}>();
</script>
<template>
	<div class="p-3 rounded-lg shadow bg-white flex flex-col gap-3">
		<SmallLabel>{{ title }}</SmallLabel>
		<slot></slot>
	</div>
</template>