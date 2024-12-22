local M = {}

function M:init()
	-- Use ALT to swap lines like vscode
	vim.keymap.set("n", "<A-j>", ":m .+1<CR>==")
	vim.keymap.set("n", "<A-k>", ":m .-2<CR>==")
	vim.keymap.set("x", "<A-j>", ":m '>+1<CR>gv-gv")
	vim.keymap.set("x", "<A-k>", ":m '<-2<CR>gv-gv")

	-- Prevents pattern match on current word from moving forward
	vim.keymap.set("n", "*", "*``")
end

return M
