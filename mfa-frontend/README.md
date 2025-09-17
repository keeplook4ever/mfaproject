
# 安装 nodejs（如果没装，可以用 brew）

brew install node

# 新建 React 项目（vite 版本更轻量）

npm create vite@latest mfa-frontend -- --template react
cd mfa-frontend

# 安装依赖

npm install

# 安装 TailwindCSS

npm install -D tailwindcss postcss autoprefixer
npx tailwindcss init -p


# 安装 yarn

npm install -g yarn
yarn add -D tailwindcss postcss autoprefixer

# 手动配置
>tailwind.config.js
> ```js
> /** @type {import('tailwindcss').Config} */
> export default {
>  content: ["./index.html", "./src/**/*.{js,ts,jsx,tsx}"],
>  theme: {
>    extend: {},
>  },
>  plugins: [],
> }



>postcss.config.js
> ``` js
> export default {
>  plugins: {
>    tailwindcss: {},
>    autoprefixer: {},
>  },
> }
