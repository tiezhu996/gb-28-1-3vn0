/** @type {import('next').NextConfig} */
const nextConfig = {
  output: 'export',
  images: { unoptimized: true },
  // 本地开发时把 /api 代理到后端
  async rewrites() {
    return [{ source: '/api/:path*', destination: 'http://localhost:3003/api/:path*' }];
  },
};

module.exports = nextConfig;
