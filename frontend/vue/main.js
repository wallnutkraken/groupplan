import Vue from 'vue'
import App from './App.vue'
import VueCookies from 'vue-cookies'
import VCalendar from 'v-calendar';

Vue.config.productionTip = false
Vue.use(VueCookies);
Vue.use(VCalendar, {
    componentPrefix: 'v',
});

Vue.mixin({
    methods: {
        // getClaims reads the groupplan JWT cookie and returns the parsed JSON claims object
        getClaims: function() {
            var jwtCookieName = 'groupplan_jwt';
            var jwtBase64 = this.$cookies.get(jwtCookieName).split('.')[1];
            var claims = JSON.parse(atob(jwtBase64));
            return claims;
        },
    },
})

/* eslint-disable no-new */
new Vue({
    el: '#app',
    components: { App },
    template: '<App/>',
})