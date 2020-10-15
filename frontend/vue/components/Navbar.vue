<template>
	<nav>
		<img class="avatar" :src="avatarURL">
		<p>{{displayName}}</p>
		<p>{{debug_cookieinfo}}</p>
	</nav>
</template>


<script>
export default {
	name: 'Navbar',
	data() {
		var jwtCookieName = 'groupplan_jwt';
		var cookies = this.$cookie.get(jwtCookieName);
		return {
			avatarURL: "",
			displayName: "",
			debug_cookieinfo: cookies,
		}
	}
}
</script>

<style scoped>
img.avatar{
	border-radius: 50%;
}
</style>
