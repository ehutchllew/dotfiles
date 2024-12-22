local M = {}

function M:init()
	-- Source current file (%)
	vim.keymap.set("n", "<space><space>x", "<cmd>source %<CR>")
	-- Source current line
	vim.keymap.set("n", "<space>x", ":.lua<CR>")
	-- Source highlighted text
	vim.keymap.set("v", "<space>x", ":lua<CR>")

	-- Remove highlight from search
	vim.keymap.set("n", "<space>h", "<cmd>noh<CR>")

	-- Navigate Quickfix Items
	vim.keymap.set("n", "<C-M-j>", "<cmd>cnext<CR>")
	vim.keymap.set("n", "<C-M-k>", "<cmd>cprev<CR>")

	-- Open Oil Navigator
	vim.keymap.set("n", "<space>e", "<cmd>Oil --float<CR>")
end

return M
