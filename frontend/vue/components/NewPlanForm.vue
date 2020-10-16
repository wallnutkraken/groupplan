<template v-if="showNewPlan == true">
	<div class="transparent-box h-100 container rounded mb-0">
		<div class="spacer">
		<div class="form-group">
			<p class="font-weight-bold">Title</p>
			<input class="form-control" id="title-input" :placeholder="pholder">
		</div>
		<div class="form-group">
			<p class="font-weight-bold">Start Date</p>
			<v-date-picker mode="range" :value="null" color="red" is-dark/>
		</div>
		</div>
	</div>
</template>

<script>
import VCalendar from 'v-calendar';
export default {
	name: 'NewPlan',
	data() {
		var claims = this.getClaims();
		return {
			pholder: claims.name + '\'s Plan',
		}
	},
	components: {
		VCalendar,
	}
}
</script>

<style scoped>
div.transparent-box{
	background-color: rgba(0,0,0,0.5);
	width: 50%;
	height: 20%;
}
div.spacer{
	margin: 30px;
}
</style>
