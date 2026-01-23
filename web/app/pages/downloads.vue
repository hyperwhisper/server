<script setup lang="ts">
import { ArrowLeft, Download, Terminal, Package, Snowflake, Copy, Check, ExternalLink } from 'lucide-vue-next'

useHead({
  title: 'Downloads - HyperWhisper',
  meta: [
    {
      name: 'description',
      content: 'Download HyperWhisper for Linux - APT, RPM, and Nix packages available',
    },
  ],
})

const copiedCommand = ref<string | null>(null)

async function copyCommand(command: string, id: string) {
  await navigator.clipboard.writeText(command)
  copiedCommand.value = id
  setTimeout(() => {
    copiedCommand.value = null
  }, 2000)
}

const debInstallCommand = 'sudo dpkg -i hyperwhisper_*.deb'
const rpmInstallCommand = 'sudo rpm -i hyperwhisper-*.rpm'

const nixBuildCommand = `git clone https://github.com/hyperwhisper/app.git
cd hyperwhisper
nix build`

const nixRunCommand = './result/bin/hyperwhisper'

const udevRulesCommand = `sudo tee /etc/udev/rules.d/99-uinput.rules << 'EOF'
KERNEL=="uinput", MODE="0660", GROUP="input", OPTIONS+="static_node=uinput"
EOF
sudo udevadm trigger --name-match=uinput`

const ydotooldServiceCommand = `mkdir -p ~/.config/systemd/user/
cat > ~/.config/systemd/user/ydotoold.service << 'EOF'
[Unit]
Description=ydotoold daemon

[Service]
ExecStart=/usr/bin/ydotoold
Restart=always

[Install]
WantedBy=default.target
EOF

# Enable and start the service
systemctl --user enable --now ydotoold.service`

const inputGroupCommand = 'sudo usermod -aG input $USER'
</script>

<template>
  <div class="min-h-screen bg-white dark:bg-black text-neutral-900 dark:text-white transition-colors duration-300">
    <!-- Subtle grain texture overlay -->
    <div class="fixed inset-0 opacity-[0.02] dark:opacity-[0.015] pointer-events-none bg-[url('data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHdpZHRoPSIzMDAiIGhlaWdodD0iMzAwIj48ZmlsdGVyIGlkPSJhIiB4PSIwIiB5PSIwIj48ZmVUdXJidWxlbmNlIGJhc2VGcmVxdWVuY3k9Ii43NSIgc3RpdGNoVGlsZXM9InN0aXRjaCIgdHlwZT0iZnJhY3RhbE5vaXNlIi8+PC9maWx0ZXI+PHJlY3Qgd2lkdGg9IjMwMCIgaGVpZ2h0PSIzMDAiIGZpbHRlcj0idXJsKCNhKSIgb3BhY2l0eT0iMSIvPjwvc3ZnPg==')]" />

    <AppNavbar />

    <main class="min-h-screen flex flex-col items-center px-4 pt-24 pb-16">
      <div class="w-full max-w-6xl">
        <!-- Header -->
        <div class="text-center mb-12">
          <div class="inline-flex items-center justify-center w-16 h-16 rounded-2xl bg-neutral-100 dark:bg-white/5 mb-6">
            <Download class="size-8 text-neutral-600 dark:text-neutral-400" />
          </div>
          <h1 class="text-3xl sm:text-4xl font-bold mb-3">Download HyperWhisper</h1>
          <p class="text-neutral-500 dark:text-neutral-400">
            Choose your preferred installation method for Linux
          </p>
        </div>

        <!-- Two Column Layout -->
        <div class="grid grid-cols-1 lg:grid-cols-2 gap-8 lg:gap-12">
          <!-- Download Options Column -->
          <div>
            <h2 class="text-xl font-semibold mb-4">Download Options</h2>
            <p class="text-sm text-neutral-500 dark:text-neutral-400 mb-6">
              Choose your preferred installation method for Linux
            </p>

            <div class="space-y-4">
              <!-- APT (Debian/Ubuntu) -->
              <div class="rounded-xl border border-neutral-200 dark:border-white/10 bg-neutral-50 dark:bg-white/[0.02] overflow-hidden">
                <div class="p-5">
                  <div class="flex items-start gap-4">
                    <div class="flex-shrink-0 w-12 h-12 rounded-lg bg-orange-100 dark:bg-orange-500/10 flex items-center justify-center">
                      <Package class="size-6 text-orange-600 dark:text-orange-400" />
                    </div>
                    <div class="flex-1 min-w-0">
                      <h3 class="text-lg font-semibold mb-1">APT Package</h3>
                      <p class="text-sm text-neutral-500 dark:text-neutral-400 mb-4">
                        For Debian, Ubuntu, Linux Mint, and other Debian-based distributions
                      </p>
                      <a
                        href="https://github.com/hyperwhisper/app/releases/latest"
                        target="_blank"
                        rel="noopener"
                        class="inline-flex items-center gap-2 px-4 py-2 rounded-lg bg-neutral-900 dark:bg-white text-white dark:text-neutral-900 hover:bg-neutral-800 dark:hover:bg-neutral-100 transition-colors text-sm font-medium"
                      >
                        <Download class="size-4" />
                        Download .deb
                        <ExternalLink class="size-3 opacity-50" />
                      </a>

                      <!-- Install command -->
                      <div class="mt-4">
                        <p class="text-xs text-neutral-500 dark:text-neutral-400 mb-2">Then install with:</p>
                        <div class="flex items-center gap-2 p-2.5 rounded-lg bg-neutral-900 dark:bg-black border border-neutral-800 dark:border-white/10 font-mono text-sm text-neutral-300">
                          <Terminal class="size-4 flex-shrink-0 text-neutral-500" />
                          <code>{{ debInstallCommand }}</code>
                          <button
                            @click="copyCommand(debInstallCommand, 'deb-install')"
                            class="ml-auto flex-shrink-0 p-1 rounded hover:bg-white/10 text-neutral-400 hover:text-white transition-colors"
                            title="Copy command"
                          >
                            <Check v-if="copiedCommand === 'deb-install'" class="size-4 text-green-400" />
                            <Copy v-else class="size-4" />
                          </button>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              </div>

              <!-- RPM (Fedora/RHEL) -->
              <div class="rounded-xl border border-neutral-200 dark:border-white/10 bg-neutral-50 dark:bg-white/[0.02] overflow-hidden">
                <div class="p-5">
                  <div class="flex items-start gap-4">
                    <div class="flex-shrink-0 w-12 h-12 rounded-lg bg-blue-100 dark:bg-blue-500/10 flex items-center justify-center">
                      <Package class="size-6 text-blue-600 dark:text-blue-400" />
                    </div>
                    <div class="flex-1 min-w-0">
                      <h3 class="text-lg font-semibold mb-1">RPM Package</h3>
                      <p class="text-sm text-neutral-500 dark:text-neutral-400 mb-4">
                        For Fedora, RHEL, CentOS, openSUSE, and other RPM-based distributions
                      </p>
                      <a
                        href="https://github.com/hyperwhisper/app/releases/latest"
                        target="_blank"
                        rel="noopener"
                        class="inline-flex items-center gap-2 px-4 py-2 rounded-lg bg-neutral-900 dark:bg-white text-white dark:text-neutral-900 hover:bg-neutral-800 dark:hover:bg-neutral-100 transition-colors text-sm font-medium"
                      >
                        <Download class="size-4" />
                        Download .rpm
                        <ExternalLink class="size-3 opacity-50" />
                      </a>

                      <!-- Install command -->
                      <div class="mt-4">
                        <p class="text-xs text-neutral-500 dark:text-neutral-400 mb-2">Then install with:</p>
                        <div class="flex items-center gap-2 p-2.5 rounded-lg bg-neutral-900 dark:bg-black border border-neutral-800 dark:border-white/10 font-mono text-sm text-neutral-300">
                          <Terminal class="size-4 flex-shrink-0 text-neutral-500" />
                          <code>{{ rpmInstallCommand }}</code>
                          <button
                            @click="copyCommand(rpmInstallCommand, 'rpm-install')"
                            class="ml-auto flex-shrink-0 p-1 rounded hover:bg-white/10 text-neutral-400 hover:text-white transition-colors"
                            title="Copy command"
                          >
                            <Check v-if="copiedCommand === 'rpm-install'" class="size-4 text-green-400" />
                            <Copy v-else class="size-4" />
                          </button>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              </div>

              <!-- Nix -->
              <div class="rounded-xl border border-neutral-200 dark:border-white/10 bg-neutral-50 dark:bg-white/[0.02] overflow-hidden">
                <div class="p-5">
                  <div class="flex items-start gap-4">
                    <div class="flex-shrink-0 w-12 h-12 rounded-lg bg-cyan-100 dark:bg-cyan-500/10 flex items-center justify-center">
                      <Snowflake class="size-6 text-cyan-600 dark:text-cyan-400" />
                    </div>
                    <div class="flex-1 min-w-0">
                      <h3 class="text-lg font-semibold mb-1">Nix / NixOS</h3>
                      <p class="text-sm text-neutral-500 dark:text-neutral-400 mb-4">
                        Build from source using Nix flakes
                      </p>
                      <!-- Build commands -->
                      <div class="relative">
                        <div class="flex items-start gap-2 p-3 rounded-lg bg-neutral-900 dark:bg-black border border-neutral-800 dark:border-white/10 font-mono text-sm text-neutral-300">
                          <Terminal class="size-4 flex-shrink-0 text-neutral-500 mt-0.5" />
                          <pre class="flex-1 whitespace-pre overflow-x-auto">{{ nixBuildCommand }}</pre>
                          <button
                            @click="copyCommand(nixBuildCommand, 'nix-build')"
                            class="flex-shrink-0 p-1.5 rounded hover:bg-white/10 text-neutral-400 hover:text-white transition-colors"
                            title="Copy command"
                          >
                            <Check v-if="copiedCommand === 'nix-build'" class="size-4 text-green-400" />
                            <Copy v-else class="size-4" />
                          </button>
                        </div>
                      </div>

                      <!-- Run command -->
                      <div class="mt-4 p-4 rounded-lg bg-emerald-50 dark:bg-emerald-500/10 border border-emerald-200 dark:border-emerald-500/20">
                        <p class="text-sm font-medium text-emerald-800 dark:text-emerald-300 mb-2">
                          After building, run with:
                        </p>
                        <div class="flex items-center gap-2 p-2.5 rounded-md bg-emerald-100 dark:bg-emerald-900/30 font-mono text-sm text-emerald-900 dark:text-emerald-200">
                          <Terminal class="size-4 flex-shrink-0 text-emerald-600 dark:text-emerald-400" />
                          <code>{{ nixRunCommand }}</code>
                          <button
                            @click="copyCommand(nixRunCommand, 'nix-run')"
                            class="ml-auto flex-shrink-0 p-1 rounded hover:bg-emerald-200 dark:hover:bg-emerald-800/50 text-emerald-600 dark:text-emerald-400 transition-colors"
                            title="Copy command"
                          >
                            <Check v-if="copiedCommand === 'nix-run'" class="size-4" />
                            <Copy v-else class="size-4" />
                          </button>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- Setup Steps for Linux Distributions -->
          <div>
            <h2 class="text-xl font-semibold mb-4">Setup Steps for Linux Distributions</h2>
            <p class="text-sm text-neutral-500 dark:text-neutral-400 mb-6">
              After installing HyperWhisper, complete these steps to enable the <span class="font-semibold text-neutral-700 dark:text-neutral-200">auto type</span> feature:
            </p>

            <div class="space-y-4">
              <!-- Step 1: udev rules -->
              <div class="rounded-xl border border-neutral-200 dark:border-white/10 bg-neutral-50 dark:bg-white/[0.02] p-5">
                <div class="flex items-start gap-3 mb-4">
                  <span class="flex-shrink-0 w-6 h-6 rounded-full bg-neutral-900 dark:bg-white text-white dark:text-neutral-900 text-xs font-bold flex items-center justify-center">1</span>
                  <div>
                    <h3 class="font-medium">Configure uinput device permissions</h3>
                    <p class="text-sm text-neutral-500 dark:text-neutral-400 mt-1">
                      Ensure <code class="px-1.5 py-0.5 rounded bg-neutral-200 dark:bg-white/10 font-mono text-xs">/dev/uinput</code> is owned by <code class="px-1.5 py-0.5 rounded bg-neutral-200 dark:bg-white/10 font-mono text-xs">root</code> user and <code class="px-1.5 py-0.5 rounded bg-neutral-200 dark:bg-white/10 font-mono text-xs">input</code> group
                    </p>
                  </div>
                </div>
                <div class="flex items-start gap-2 p-3 rounded-lg bg-neutral-900 dark:bg-black border border-neutral-800 dark:border-white/10 font-mono text-sm text-neutral-300">
                  <Terminal class="size-4 flex-shrink-0 text-neutral-500 mt-0.5" />
                  <pre class="flex-1 whitespace-pre overflow-x-auto">{{ udevRulesCommand }}</pre>
                  <button
                    @click="copyCommand(udevRulesCommand, 'udev-rules')"
                    class="flex-shrink-0 p-1.5 rounded hover:bg-white/10 text-neutral-400 hover:text-white transition-colors"
                    title="Copy command"
                  >
                    <Check v-if="copiedCommand === 'udev-rules'" class="size-4 text-green-400" />
                    <Copy v-else class="size-4" />
                  </button>
                </div>
              </div>

              <!-- Step 2: ydotoold service -->
              <div class="rounded-xl border border-neutral-200 dark:border-white/10 bg-neutral-50 dark:bg-white/[0.02] p-5">
                <div class="flex items-start gap-3 mb-4">
                  <span class="flex-shrink-0 w-6 h-6 rounded-full bg-neutral-900 dark:bg-white text-white dark:text-neutral-900 text-xs font-bold flex items-center justify-center">2</span>
                  <div>
                    <h3 class="font-medium">Create and enable ydotoold user service</h3>
                    <p class="text-sm text-neutral-500 dark:text-neutral-400 mt-1">
                      Set up <code class="px-1.5 py-0.5 rounded bg-neutral-200 dark:bg-white/10 font-mono text-xs">ydotoold</code> as a user service that starts automatically
                    </p>
                  </div>
                </div>
                <div class="flex items-start gap-2 p-3 rounded-lg bg-neutral-900 dark:bg-black border border-neutral-800 dark:border-white/10 font-mono text-sm text-neutral-300">
                  <Terminal class="size-4 flex-shrink-0 text-neutral-500 mt-0.5" />
                  <pre class="flex-1 whitespace-pre overflow-x-auto">{{ ydotooldServiceCommand }}</pre>
                  <button
                    @click="copyCommand(ydotooldServiceCommand, 'ydotoold-service')"
                    class="flex-shrink-0 p-1.5 rounded hover:bg-white/10 text-neutral-400 hover:text-white transition-colors"
                    title="Copy command"
                  >
                    <Check v-if="copiedCommand === 'ydotoold-service'" class="size-4 text-green-400" />
                    <Copy v-else class="size-4" />
                  </button>
                </div>
              </div>

              <!-- Step 3: input group -->
              <div class="rounded-xl border border-neutral-200 dark:border-white/10 bg-neutral-50 dark:bg-white/[0.02] p-5">
                <div class="flex items-start gap-3 mb-4">
                  <span class="flex-shrink-0 w-6 h-6 rounded-full bg-neutral-900 dark:bg-white text-white dark:text-neutral-900 text-xs font-bold flex items-center justify-center">3</span>
                  <div>
                    <h3 class="font-medium">Add your user to the input group</h3>
                    <p class="text-sm text-neutral-500 dark:text-neutral-400 mt-1">
                      You'll need to log out and back in for this to take effect
                    </p>
                  </div>
                </div>
                <div class="flex items-center gap-2 p-2.5 rounded-lg bg-neutral-900 dark:bg-black border border-neutral-800 dark:border-white/10 font-mono text-sm text-neutral-300">
                  <Terminal class="size-4 flex-shrink-0 text-neutral-500" />
                  <code>{{ inputGroupCommand }}</code>
                  <button
                    @click="copyCommand(inputGroupCommand, 'input-group')"
                    class="ml-auto flex-shrink-0 p-1 rounded hover:bg-white/10 text-neutral-400 hover:text-white transition-colors"
                    title="Copy command"
                  >
                    <Check v-if="copiedCommand === 'input-group'" class="size-4 text-green-400" />
                    <Copy v-else class="size-4" />
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Back Button -->
        <div class="mt-12 flex justify-center">
          <Button
            variant="ghost"
            class="text-neutral-500 dark:text-neutral-400 hover:text-neutral-900 dark:hover:text-white hover:bg-neutral-100 dark:hover:bg-white/5 gap-2"
            as-child
          >
            <NuxtLink to="/">
              <ArrowLeft class="size-4" />
              Back to Home
            </NuxtLink>
          </Button>
        </div>
      </div>
    </main>
  </div>
</template>
