-- Set <space> as leader key
vim.g.mapleader = " "
vim.g.maplocalleader = " "

-- Set to true if Nerd Font used in terminal
vim.g.have_nerd_font = true
vim.cmd "colorscheme desert"

-- ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
-- [[ Keymaps ]]
-- For list of options run: `:help vim.keymap.set()`
-- ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
-- ~~~~~
-- [[ Normal Mode ]]
-- ~~~~~
--
-- Movement:
vim.keymap.set('n', "<A-j>", ":m .+1<CR>==", { desc = "Move current line/block down (vscode)" })
vim.keymap.set('n', "<A-k>", ":m .-2<CR>==", { desc = "Move current line/block up (vscode)" })

vim.keymap.set('n', "<C-h>", "<C-w>h", { desc = "Jump to leftward window" })
vim.keymap.set('n', "<C-j>", "<C-w>j", { desc = "Jump to upward window" })
vim.keymap.set('n', "<C-k>", "<C-w>k", { desc = "Jump to downward window" })
vim.keymap.set('n', "<C-l>", "<C-w>l", { desc = "Jump to rightward window" })

-- ~~~~~
-- [[ Visual Block Mode ]]
-- ~~~~~
--
-- Movement:
vim.keymap.set('x', "<A-j>", ":m '>+1<CR>gv-gv", { desc = "Move current line/block down (vscode)" })
vim.keymap.set('x', "<A-k>", ":m '<-2<CR>gv-gv", { desc = "Move current line/block up (vscode)" })



-- ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
-- [[ Setting Options ]]
-- For list of options run: `:help option-list`
-- ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

-- Indent wrapped lines to be on same indent level
vim.opt.breakindent = true

-- Sync clipboard between OS and Neovim
-- See `:help 'clipboard'`
vim.opt.clipboard = "unnamedplus"

-- Show popup from completion
vim.opt.completeopt = { "menuone", "noselect" }
vim.opt.pumheight = 10

-- Show which line your cursor is on
vim.opt.cursorline = true

-- Enforce indents to be 2 spaces
vim.opt.expandtab = true
vim.opt.shiftwidth = 2
vim.opt.tabstop = 2

-- Scope Folding
-- vim.opt.foldmethod = "expr"
-- vim.opt.foldexpr = "nvim_treesitter#foldexpr()"

-- Preview substitutions live, as you type!
vim.opt.inccommand = 'split'

-- Display certain whitespace
vim.opt.list = true
vim.opt.listchars = { tab = '» ', trail = '·', nbsp = '␣' }

-- Enable mouse mode, can be useful for resizing splits
vim.opt.mouse = "a"

-- Make line numbers default
vim.opt.number = true
vim.opt.relativenumber = true

-- Minimal number of screen lines to keep above, below, and to the sides the cursor.
vim.opt.scrolloff = 10
vim.opt.sidescrolloff = 10

-- Don't show mode, will be on status line
vim.opt.showmode = false

-- Always show sign column
vim.opt.signcolumn = "yes"

-- Set term gui colors
vim.opt.termguicolors = true

-- Set window title
vim.opt.title = true

-- Save undo history
vim.opt.undofile = true



-- ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
-- [[ Basic Autocommands (event listeners) ]]
-- For list of options run: `:help lua-guide-autocommands`
-- ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

-- Highlight when yanking (copying) text
-- See `:help vim.highlight.on_yank()`
vim.api.nvim_create_autocmd('TextYankPost', {
	desc = "Highlight when yanking (copying) text",
	group = vim.api.nvim_create_augroup("kickstart-highlight-yank", { clear = true }),
	callback = function()
		vim.highlight.on_yank()
	end,
})

-- Install `lazy.nvim` plugin manager
-- See `:help lazy.nvim.txt` or https://github.com/folke/lazy.nvim
local lazypath = vim.fn.stdpath "data" .. "/lazy/lazy.nvim"
if not vim.loop.fs_stat(lazypath) then
	local lazyrepo = "https://github.com/folke/lazy.nvim.git"
	vim.fn.system { "git", "clone", "--filter=blob:none", "--branch=stable", lazyrepo, lazypath }
end ---@diagnostic disable-next-line: undefined-field
vim.opt.rtp:prepend(lazypath)


-- ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
-- [[ Basic Autocommands (event listeners) ]]
-- For list of options run: `:help lua-guide-autocommands`
-- ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~



-- ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
-- [[ Plugins: Configure & Install ]]
--
-- To check current status of plugins run: `:Lazy`
-- Press `?` in menu for help. Use `:q` to close
-- To update plugins run: `:Lazy update`
--
-- NOTE: Plugins can be added with a link or `owner/repo`
-- ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

-- NOTE: Plugins can run lua code when they are loaded
-- Useful for group configs and lazy loading

require("lazy").setup({
	-- Autocompletion
	{
		"hrsh7th/nvim-cmp",
		event = "InsertEnter",
		dependencies = {
			"cmp-buffer",
			"cmp-cmdline",
			"cmp-nvim-lsp",
			"cmp-path"
		}
	},
	{ "hrsh7th/cmp-buffer", lazy = true },
	{ "hrsh7th/cmp-cmdline", lazy = true },
	{ "hrsh7th/cmp-nvim-lsp", lazy = true },
	{ "hrsh7th/cmp-path", lazy = true },
	-- Colorschemes
	{ "ellisonleao/gruvbox.nvim" },
	-- Comment
	{
		"numToStr/Comment.nvim",
		opts = {}
	},
	-- LSP
	{
		"neovim/nvim-lspconfig",
		dependencies = {
			{
				"williamboman/mason.nvim",
				config = true
			},
			"williamboman/mason-lspconfig.nvim",
			{ "folke/neodev.nvim", opts = {} }
		},
		config = function()
			require("plugins.lsp").setup()
		end
	},
	-- Telescope
	{
		"nvim-telescope/telescope.nvim",
		event = "VimEnter",
		branch = "0.1.x",
		cmd = "Telescope",
		dependencies = {
			"nvim-lua/plenary.nvim",
			{
				"nvim-telescope/telescope-fzf-native.nvim",
				build = "make"
			},
			{
				"nvim-telescope/telescope-ui-select.nvim"
			},
			{
				"nvim-tree/nvim-web-devicons",
				enabled = vim.g.have_nerd_font
			}
		},
		config = function()
			require("plugins.telescope").setup()
		end
	},
	-- WhichKey
	{
		"folke/which-key.nvim",
		event = "VeryLazy",
		config = function()
			require("plugins.which-key").setup()
		end
	},
})
