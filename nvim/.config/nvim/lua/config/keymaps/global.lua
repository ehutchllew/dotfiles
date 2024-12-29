local M = {}

function M:init()
	-- Source current file (%)
	vim.keymap.set("n", "<space><space>x", "<cmd>source %<CR>", { desc = "Source Current File" })
	-- Source current line
	vim.keymap.set("n", "<space>x", ":.lua<CR>", { desc = "Source Current Line" })
	-- Source highlighted text
	vim.keymap.set("v", "<space>x", ":lua<CR>", { desc = "Source Highlighted Text" })

	-- Remove highlight from search
	vim.keymap.set("n", "<space>h", "<cmd>noh<CR>", { desc = "Remove Highlight from Search" })

	-- Navigate Quickfix Items
	vim.keymap.set("n", "<C-M-j>", "<cmd>cnext<CR>", { desc = "Move to Next Quickfix Item" })
	vim.keymap.set("n", "<C-M-k>", "<cmd>cprev<CR>", { desc = "Move to Previous Quickfix Item" })

	-- Open Oil Navigator
	vim.keymap.set("n", "<space>e", "<cmd>Oil --float<CR>", { desc = "Open Oil Navigator" })

	-- Toggle Neovim Floating Terminal
	vim.keymap.set("n", "<C-\\>", ":Floaterminal<CR>", { desc = "Toggle Floating Terminal" })
	--vim.keymap.set("t", "<C-\\>", "<C-\\><C-n> :lua Floaterminal<CR>")
end

return M
