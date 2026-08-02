import js from '@eslint/js'
import svelte from 'eslint-plugin-svelte'
import globals from 'globals'
import tseslint from 'typescript-eslint'

export default tseslint.config(
  js.configs.recommended,
  ...tseslint.configs.recommended,
  ...svelte.configs['flat/recommended'],
  {
    files: ['**/*.svelte', '**/*.svelte.js', '**/*.svelte.ts'],
    languageOptions: {
      parserOptions: {
        parser: tseslint.parser,
      },
    },
  },
  {
    languageOptions: {
      globals: {
        ...globals.browser,
        ...globals.node,
      },
    },
  },
  {
    ignores: ['build/', '.svelte-kit/', 'dist/', 'node_modules/'],
  },
)
