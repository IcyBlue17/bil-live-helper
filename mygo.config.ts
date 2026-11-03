import { defineConfig } from "mygo-cli";
import pkg from "./package.json" with { type: "json" };

export default defineConfig({
  // No spaces: the Windows build is a single portable BilLiveHelper.exe,
  // uploaded to releases under this exact name (GitHub turns spaces in
  // asset names into dots), and updates replace the file of this name.
  // main.go sets the name users see back to "Bil Live Helper".
  name: "BilLiveHelper",
  slug: "bil-live-helper",
  identifier: "dev.icyblue.billivehelper",
  // The release workflow checks that the tag is v + this version.
  version: pkg.version,
  // `mygo dev` runs devCommand and loads devUrl; `mygo build` runs
  // buildCommand and embeds frontendDist into the app.
  devUrl: "http://localhost:5173",
  devCommand: "bun run dev:web",
  buildCommand: "bun run build:web",
  frontendDist: "dist",
  // Outside resources/, which would ship next to the executable.
  icon: "assets/icon.png",
  bindings: "src/mygo.ts",
  out: "build",
  // Installed apps update from the latest GitHub release, signed with the
  // key in the MYGO_UPDATER_PRIVATE_KEY secret.
  updates: {
    publicKey: "a3ODtaYIphviOqJx8/HsZVcYUJ/u1ovjMh3pZnqvvqk=",
    github: "IcyBlue17/bil-live-helper",
  },
  windows: { displayName: "Bil Live Helper" },
  macos: { infoPlist: { CFBundleDisplayName: "Bil Live Helper" } },
});
