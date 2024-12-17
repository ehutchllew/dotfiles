return {
  {
    "echasnovski/mini.nvim",
    enabled = true,
    opts = {
      use_icons = true
    },
    config = function(_, opts)
      local statusline = require("mini.statusline")
      statusline.setup(opts)
    end
  }
}
