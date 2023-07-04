-- Read the docs: https://www.lunarvim.org/docs/configuration
-- Video Tutorials: https://www.youtube.com/watch?v=sFA9kX-Ud_c&list=PLhoH5vyxr6QqGu0i7tt_XoVK9v-KvZ3m6
-- Forum: https://www.reddit.com/r/lunarvim/
-- Discord: https://discord.com/invite/Xb9B4Ny

-- Uses relative line numbers
vim.opt.rnu = true

--
-- MAPPINGS
--

-- Prevents match operator from jumping to next word
lvim.keys.normal_mode["*"] = "*``"
-- Creates new map `gt` specifically for going to type definition
lvim.keys.normal_mode["gt"] = { vim.lsp.buf.type_definition, { desc = "Goto Type Definition" } }
-- Add mapping for MiniMap toggle
lvim.keys.normal_mode["mm"] = { "<cmd>lua MiniMap.toggle() <cr>", { desc = "Toggle MiniMap" } }
-- Utilizes Telescope popup for references instead of native vim refs
lvim.lsp.buffer_mappings.normal_mode["gr"][1] = "<cmd>Telescope lsp_references <cr>"

--
-- EDITOR
--

-- Adds filename to statusline
lvim.builtin.lualine.sections.lualine_c = { "diff", "filename" }
lvim.colorscheme = "catppuccin-mocha"
lvim.format_on_save = true


--
-- PLUGINS
--

lvim.plugins = {
  {
    "nvim-treesitter/nvim-treesitter-context"
  },
  {
    "catppuccin/nvim",
    name = "catppuccin",
    priority = 1000
  },
  {
    'echasnovski/mini.map',
    config = function()
      local map = require("mini.map")
      map.setup({
        symbols = {
          encode = map.gen_encode_symbols.dot('4x2')
        }
      })
    end,
    version = '*'
  },
  {
    'echasnovski/mini.splitjoin',
    config = function()
      require("mini.splitjoin").setup()
    end,
    version = '*'
  },
  {
    'echasnovski/mini.surround',
    config = function()
      require("mini.surround").setup()
    end,
    version = '*'
  },
  {
    "folke/trouble.nvim",
    dependencies = { "nvim-tree/nvim-web-devicons" },
    opts = {}
  }
}
