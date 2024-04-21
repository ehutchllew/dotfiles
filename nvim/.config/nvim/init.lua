-- Set <space> as leader key
vim.g.mapleader = " "
vim.g.maplocalleader = " "

-- Set to true if Nerd Font used in terminal
vim.g.have_nerd_font = true

-- ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
-- [[ Setting Options ]]
-- For list of options run: `:help option-list`
-- ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

-- Indent wrapped lines to be on same indent level
vim.opt.breakindent = true

-- Sync clipboard between OS and Neovim
-- See `:help 'clipboard'`
vim.opt.clipboard = "unnamedplus"

-- Enable mouse mode, can be useful for resizing splits
vim.opt.mouse = "a"

-- Make line numbers default
vim.opt.number = true
vim.opt.relativenumber = true

-- Don't show mode, will be on status line
vim.opt.showmode = false

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
	{
		"folke/which-key.nvim",
		event = "VeryLazy",
		config = function()
			require("plugins.which-key").setup()
		end
	},
})
