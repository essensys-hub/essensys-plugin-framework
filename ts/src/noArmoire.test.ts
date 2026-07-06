import { describe, it, expect } from "vitest";
import { assertNoArmoire, ArmoireMutationBlocked } from "./noArmoire";

describe("no-armoire guard", () => {
  it("bloque une mutation vers l'armoire", () => {
    expect(() => assertNoArmoire("/api/admin/inject", "POST")).toThrow(ArmoireMutationBlocked);
    expect(() => assertNoArmoire("/scenarios/42/launch", "POST")).toThrow(ArmoireMutationBlocked);
  });

  it("laisse passer une lecture", () => {
    expect(() => assertNoArmoire("/api/plugins/sungrow-solar/current", "GET")).not.toThrow();
  });

  it("laisse passer une mutation non-armoire", () => {
    expect(() => assertNoArmoire("/api/plugins/sungrow-solar/ack", "POST")).not.toThrow();
  });

  it("autorise le dry-run explicite", () => {
    expect(() => assertNoArmoire("/api/admin/inject", "POST", true)).not.toThrow();
  });
});
