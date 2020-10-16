<template>
<div id="app">
  <Navbar/>
  <NewPlanForm />
  <!-- v-if="showNewPlan == true" -->
</div>
</template>

<script>
import Navbar from './components/Navbar.vue'
import NewPlanForm from "./components/NewPlanForm.vue"

export default {
  name: 'App',
  components: {
    Navbar,
    NewPlanForm,
  },
  data() {
    return {
      showNewPlan: false,
    }
  },
  
  methods: {
    enablenewplan: function(event) {
      if (this.showNewPlan == true) {
        // TODO: reset the plan
        return;
      }
      this.showNewPlan = true;
    },
  },
}
</script>

<style>

#app {
  font-family: 'Avenir', Helvetica, Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  text-align: center;
  color: #2c3e50;
  margin-top: 60px;
}
</style>
