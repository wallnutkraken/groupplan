<template>
<nav class="navbar fixed-top navbar-light bg-light">
  <a class="navbar-brand">
    <img :src="avatarURL" width="30" height="30" class="d-inline-block align-top avatar" alt="">
    {{displayName}}
  </a>
  <button type="button" v-on:click="enablenewplan" class="btn btn-success">New Plan</button>
</nav>
</template>


<script>

export default {
	name: 'Navbar',
	data() {
		// Get the JWT claims object, it should contain the user's display name and profile picture
		var claims = this.getClaims();
		return {
			avatarURL: claims.pfp,
			displayName: claims.name,
		}
	}
}
</script>

<style scoped>
img.avatar{
	border-radius: 50%;
}
</style>
