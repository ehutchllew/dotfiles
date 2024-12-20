-- Read the docs: https://www.lunarvim.org/docs/configuration
-- Video Tutorials: https://www.youtube.com/watch?v=sFA9kX-Ud_c&list=PLhoH5vyxr6QqGu0i7tt_XoVK9v-KvZ3m6
-- Forum: https://www.reddit.com/r/lunarvim/
-- Discord: https://discord.com/invite/Xb9B4Ny

-- Uses relative line numbers
vim.opt.rnu = true
lvim.builtin.nvimtree.setup.view.relativenumber = true

-- Folding config
vim.opt.foldmethod = "indent"
vim.opt.foldlevel = 99

lvim.builtin.project.detection_methods = { "lsp", "pattern" }
lvim.builtin.project.patterns = {
  ".git",
  "go.mod",
  "go.work",
  "package.json",
  "deno.json",
  "deno.jsonc",
}

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
lvim.colorscheme = "gruvbox"
lvim.format_on_save = true

--
-- LSP
--
vim.list_extend(lvim.lsp.automatic_configuration.skipped_servers, { "tsserver" })
local nvim_lsp = require("lspconfig")

local denoopts = {
  server = {
    root_dir = nvim_lsp.util.root_pattern("deno.json", "deno.jsonc")
  }
}

local tsopts = {
  server = {
    root_dir = nvim_lsp.util.root_pattern("package.json"),
    single_file_support = false
  }
}

local lvim_manager = require("lvim.lsp.manager")
lvim_manager.setup("denols", denoopts)
lvim_manager.setup("tsserver", tsopts)

--
-- FORMATTERS
--

local formatters = require "lvim.lsp.null-ls.formatters"
formatters.setup {
  {
    name = "prettierd",
    filetypes = { "typescript" }
  }
}

--
-- LINTERS
--

local linters = require "lvim.lsp.null-ls.linters"
linters.setup {
  {
    name = "eslint_d",
    filetypes = { "typescript" }
  }
}

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
    "ellisonleao/gruvbox.nvim",
    name = "gruvbox",
    priority = 1001
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
