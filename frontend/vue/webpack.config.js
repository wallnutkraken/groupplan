const HtmlWebpackPlugin = require('html-webpack-plugin');
const VueLoaderPlugin = require('vue-loader/lib/plugin');

module.exports = {
    entry: './main.js',
    resolve: {
        alias: {
            vue: 'vue/dist/vue.js'
        },
    },
    module: {
        rules: [
            { test: /\.js$/, use: 'babel-loader' },
            { test: /\.vue$/, use: 'vue-loader' },
            { test: /\.css$/, use: ['vue-style-loader', 'css-loader'] },
        ]
    },
    plugins: [
        new HtmlWebpackPlugin({
            template: '../dashboard.html',
            filename: '../debugdash.html',
        }),
        new VueLoaderPlugin(),
    ],
    output: {
        filename: 'groupplan.js',
        path: __dirname + "/../static/js"
    }
};