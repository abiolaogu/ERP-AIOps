import type { CodegenConfig } from "@graphql-codegen/cli";

const config: CodegenConfig = {
  schema: [
    {
      [`${process.env.VITE_HASURA_URL ?? 'http://localhost:19109/v1/graphql'}`]: {
        headers: {
          "X-Hasura-Admin-Secret": `${process.env.HASURA_ADMIN_SECRET ?? 'hasura-admin-secret'}`,
        },
      },
    },
  ],
  documents: ["src/graphql/**/*.graphql"],
  generates: {
    "src/graphql/generated.ts": {
      plugins: ["typescript", "typescript-operations"],
    },
  },
};

export default config;
