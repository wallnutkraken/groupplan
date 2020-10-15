<template>
  <div class="hello">
    <h1>{{ msg }}</h1>
    <h2>I need help</h2>
    <h2>How do I vue, my css <strike>doesnt work</strike></h2>
  </div>
</template>

<script>
export default {
  name: 'HelloWorld',
  data() {
    return {
      msg: 'GroupPlan? More like how do I frontend?'
    }
  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style>
body{
	background-color: #42b983;
}
h1,
h2 {
  font-weight: normal;
}
ul {
  list-style-type: none;
  padding: 0;
}
li {
  display: inline-block;
  margin: 0 10px;
}
a {
  color: #42b983;
}
</style>
