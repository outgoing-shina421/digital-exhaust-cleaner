const path = require('path');
const MiniCssExtractPlugin = require('mini-css-extract-plugin');
const CssMinimizerPlugin   = require('css-minimizer-webpack-plugin');
const TerserPlugin         = require('terser-webpack-plugin');

/** Output lands in the Go embed directory so the binary self-contains all assets. */
const OUT_DIR = path.resolve(__dirname, '../internal/ui/static');

module.exports = (env, argv) => {
  const isProd = argv.mode === 'production';

  return {
    entry: './src/scripts/app.js',

    output: {
      path:     OUT_DIR,
      filename: 'app.js',
      clean:    true,
    },

    module: {
      rules: [
        {
          test: /\.scss$/,
          use: [
            MiniCssExtractPlugin.loader,
            { loader: 'css-loader', options: { sourceMap: !isProd } },
            { loader: 'sass-loader', options: { sourceMap: !isProd, api: 'modern' } },
          ],
        },
      ],
    },

    plugins: [
      new MiniCssExtractPlugin({ filename: 'app.css' }),
    ],

    optimization: {
      minimizer: [
        new TerserPlugin({ extractComments: false }),
        new CssMinimizerPlugin(),
      ],
    },

    devtool: isProd ? false : 'source-map',
  };
};
