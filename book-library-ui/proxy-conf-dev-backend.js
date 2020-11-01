getProp = (name, defaultValue) => {
  if (process.env[name]) {
    return process.env[name];
  }

  return defaultValue;
};

getFintechUrl = () => {
  return getProp('BOOK_LIBRARY_URL', 'https://books-api-book-library.cloud.adorsys.de');
};

const PROXY_CONFIG = {
  '/book-library-api-proxy/*': {
    target: getFintechUrl(),
    pathRewrite: {
      '^/book-library-api-proxy': ''
    },
    logLevel: 'debug',
    secure: false,
    changeOrigin: true,
    headers: { 'Access-Control-Expose-Headers': '*' },
    onProxyRes: (proxyRes, req, res) => {
      if (proxyRes.headers['location']) {
        // change remote UI to local UI
        if (proxyRes.headers['location'].includes(getConsentUiUrl())) {
          proxyRes.headers['location'] = proxyRes.headers['location'].replace(
            'http://localhost:4000'
          );
        }
      }
    }
  }
};

module.exports = PROXY_CONFIG;
