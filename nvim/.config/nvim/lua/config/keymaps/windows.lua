local M = {}

function M:init()
	-- Window Navigation
	vim.keymap.set("n", "<C-h>", "<C-w>h", { desc = "Move to Left Window" })
	vim.keymap.set("n", "<C-l>", "<C-w>l", { desc = "Move to Right Window" })
	vim.keymap.set("n", "<C-j>", "<C-w>j", { desc = "Move to Below Window" })
	vim.keymap.set("n", "<C-k>", "<C-w>k", { desc = "Move to Above Window" })
end

return M
