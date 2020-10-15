import Vue from 'vue'
import App from './App.vue'
import VueCookies from 'vue-cookies'

Vue.config.productionTip = false
Vue.use(VueCookies);

/* eslint-disable no-new */
new Vue({
    el: '#app',
    components: { App },
    template: '<App/>',
})