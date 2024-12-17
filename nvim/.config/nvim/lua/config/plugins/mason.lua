return {
  {
    "williamboman/mason.nvim",
    enabled = true,
    cmd = { "Mason", "MasonInstall", "MasonUninstall", "MasonUninstallAll", "MasonLog" },
    lazy = true,
    opts = {
      ui = {
        icons = {
          package_installed = "✓",
          package_pending = "➜",
          package_uninstalled = "✗"
        }
      }
    },
    build = function()
      pcall(function()
        require("mason-registry").refresh()
      end)
    end,
    config = function(_, opts)
      local msn = require("mason")
      msn.setup(opts)
    end,
  }
}
