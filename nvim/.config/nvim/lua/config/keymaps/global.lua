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
end

return M
