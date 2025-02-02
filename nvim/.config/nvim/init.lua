--
-- OPTIONS
--

vim.opt.clipboard = "unnamedplus"
vim.opt.foldcolumn = "1"
vim.opt.foldlevel = 99
vim.opt.foldmethod = "indent"
vim.opt.number = true
vim.opt.relativenumber = true
vim.opt.shiftwidth = 2
vim.opt.tabstop = 2
vim.opt.termguicolors = true

--
-- KEYMAPS
--
require("config.keymappings"):init()

--
-- AUTO COMMANDS
--

-- Highlight when yanking (copying) text
-- Try with `yap` in normal mode
-- See `:help vim.highlight.on_yank()`
vim.api.nvim_create_autocmd('TextYankPost', {
	desc = "Highlight when yanking (copying) text",
	group = vim.api.nvim_create_augroup("kickstart-highlight-yank", { clear = true }),
	callback = function()
		vim.highlight.on_yank()
	end,
})

vim.api.nvim_create_autocmd("TermOpen", {
	group = vim.api.nvim_create_augroup("custom-term-open", { clear = true }),
	callback = function()
		vim.opt.number = false
		vim.opt.relativenumber = false
	end,
})

-- Plugins
require("config.lazy")

-- Docs
--
-- After Dir
-- Neovim runs anything in this dir after the default config
--
-- Autocmds
-- See `:lua =vim.api.nvim_create_autocmd`
-- See `:help nvim_create_autocmd()`
--
-- Autocmd Groups
-- See `:help nvim_create_augroup()`
--
-- Ftplugin
-- Stands for File Type Plugin (filetype == language)
-- Able to set options, and more, for specific file types
-- ex lua.lua or go.lua
