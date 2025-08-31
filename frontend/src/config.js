const config = {
  api: {
    baseUrl: process.env.REACT_APP_API_BASE_URL || 'http://localhost:8080',
    version: process.env.REACT_APP_API_VERSION || 'v1',
    endpoints: {
      analyze: '/api/v1/analyze',
      analysis: '/api/v1/analysis',
      health: '/health'
    }
  },
  polling: {
    interval: parseInt(process.env.REACT_APP_POLLING_INTERVAL) || 2000,
    maxAttempts: parseInt(process.env.REACT_APP_MAX_POLLING_ATTEMPTS) || 150
  }
};

export default config;
