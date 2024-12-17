return {
  {
    "williamboman/mason.nvim",
    enabled = true,
    opts = {
      ui = {
        icons = {
          package_installed = "✓",
          package_pending = "➜",
          package_uninstalled = "✗"
        }
      }
    },
    config = function(_, opts)
      local msn = require("mason")
      msn.setup(opts)
    end,
  }
}
