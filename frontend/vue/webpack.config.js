const HtmlWebpackPlugin = require('html-webpack-plugin');
const VueLoaderPlugin = require('vue-loader/lib/plugin');

module.exports = {
    entry: './main.js',

    module: {
        rules: [
            { test: /\.js$/, use: 'babel-loader' },
            { test: /\.vue$/, use: 'vue-loader' },
            { test: /\.css$/, use: ['vue-style-loader', 'css-loader'] },
        ]
    },
    plugins: [
        new VueLoaderPlugin(),
    ],
    output: {
        filename: 'groupplan.js',
        path: __dirname + "/../static/js"
    }
};