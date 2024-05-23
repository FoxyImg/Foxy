import {ref, type Ref, watch} from "vue";

export default function useImageLoader(url: Ref<string|null>) {
	const isLoading = ref(false);
	const isLoaded = ref(false);
	const error = ref<string|null>(null);
	const size = ref<{width: number, height: number}>({width: 0, height: 0});
	const loadTime = ref(0);

	watch(url, (newUrl) => {
		isLoaded.value = false;
		error.value = null;
		size.value = {width: 0, height: 0};
		loadTime.value = 0;

		if (!newUrl) {
			isLoading.value = false;
			return;
		}

		const startTime = performance.now();
		const image = new Image();
		image.onerror = () => {
			error.value = "Image could not be loaded";
			isLoading.value = false;
			isLoading.value = false;
		};

		image.onload = () => {
			isLoading.value = false;
			isLoaded.value = true;
			size.value = {width: image.width, height: image.height};
			loadTime.value = performance.now() - startTime;
		};

		isLoading.value = true;
		image.src = newUrl;
	}, {immediate: true});

	return {
		isLoading,
		isLoaded,
		error,
		size,
		loadTime,
	}
}