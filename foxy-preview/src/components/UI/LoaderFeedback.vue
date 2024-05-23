<template>
	<div class="loader"></div>
</template>
<style scoped>
/* HTML: <div class="loader"></div> */
/* HTML: <div class="loader"></div> */
.loader {
	width: 50px;
	aspect-ratio: 1;
	transform: translate(-100%,100%);
	animation: l14 1.66s infinite;
}
@keyframes l14 {
	0%{
		box-shadow:
			50px -50px #B5AC01,
			50px -50px #ECBA09,
			50px -50px #0000,
			50px -50px #0000,
			50px -50px #0000}
	25%{
		box-shadow:
			100px -50px #0000,
			50px  -50px #ECBA09,
			50px  -50px #E86E1C,
			50px  -50px #0000,
			50px  -50px #0000}
	50%{
		box-shadow:
			100px -50px  #0000,
			50px  -100px #0000,
			50px  -50px  #E86E1C,
			50px  -50px  #D41E45,
			50px  -50px  #0000}
	75%{
		box-shadow:
			100px -50px  #0000,
			50px  -100px #0000,
			0px  -50px  #0000,
			50px  -50px  #D41E45,
			50px  -50px  #B5AC01}
	100%{
		box-shadow:
			100px -50px  #0000,
			50px  -100px #0000,
			0px  -50px  #0000,
			50px    0px  #0000,
			50px  -50px  #B5AC01}
}
</style>