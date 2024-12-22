--
-- OPTIONS
--

vim.opt.clipboard = "unnamedplus"
vim.opt.number = true
vim.opt.relativenumber = true
vim.opt.shiftwidth = 2
vim.opt.tabstop = 2

--
-- KEYMAPS
--
vim.keymap.set("n", "<space><space>x", "<cmd>source %<CR>")
vim.keymap.set("n", "<space>x", ":.lua<CR>")
vim.keymap.set("v", "<space>x", ":lua<CR>")

-- Window Navigation
vim.keymap.set("n", "<C-h>", "<C-w>h")
vim.keymap.set("n", "<C-l>", "<C-w>l")
vim.keymap.set("n", "<C-j>", "<C-w>j")
vim.keymap.set("n", "<C-k>", "<C-w>k")

-- File Nav & Manipulation
-- Use ALT to swap lines like vscode
vim.keymap.set("n", "<A-j>", ":m .+1<CR>==")
vim.keymap.set("n", "<A-k>", ":m .-2<CR>==")
vim.keymap.set("x", "<A-j>", ":m '>+1<CR>gv-gv")
vim.keymap.set("x", "<A-k>", ":m '<-2<CR>gv-gv")

-- Editor Commands
-- Remove highlight from search
vim.keymap.set("n", "<space>h", "<cmd>noh<CR>")
-- Prevents pattern match on current word from moving forward
vim.keymap.set("n", "*", "*``")

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
