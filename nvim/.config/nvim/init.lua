-- Options

vim.opt.clipboard = "unnamedplus"
vim.opt.number = true
vim.opt.relativenumber = true
vim.opt.shiftwidth = 4

-- Keymaps
vim.keymap.set("n", "<space><space>x", "<cmd>source %<CR>")
vim.keymap.set("n", "<space>x", ":.lua<CR>")
vim.keymap.set("v", "<space>x", ":lua<CR>")

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
