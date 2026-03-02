import { env, type AuthPolicy } from "@/lib/config/env";

export const modulePolicy = {
  name: "ERP-AIOps",
  slug: "aiops",
  authPolicy: env.authPolicy as AuthPolicy,
  notes: "",
  altUrls: "",
};
