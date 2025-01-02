local M = {}

local get_diag = function()
	local float = vim.diagnostic.config().float

	if float then
		local config = type(float) == "table" and float or {}
		config.scope = "line"

		vim.diagnostic.open_float(config)
	end
end

function M:init()
	-- Use ALT to swap lines like vscode
	vim.keymap.set("n", "<A-j>", ":m .+1<CR>==", { desc = "Swap Current Line Below" })
	vim.keymap.set("n", "<A-k>", ":m .-2<CR>==", { desc = "Swap Current Line Above" })
	vim.keymap.set("x", "<A-j>", ":m '>+1<CR>gv-gv", { desc = "Swap Highlighted Lines Below" })
	vim.keymap.set("x", "<A-k>", ":m '<-2<CR>gv-gv", { desc = "Swap Highlighted Lines Above" })

	-- Prevents pattern match on current word from moving forward
	vim.keymap.set("n", "*", "*``", { desc = "Highlight All Matches at Cursor" })

	-- NOTE: Using telescope versions of these and defining within the telescope plugin file
	-- Go To
	--	vim.keymap.set("n", "gd", "<cmd>lua vim.lsp.buf.definition()<cr>", { desc = "GoTo Definition" })
	vim.keymap.set("n", "gD", "<cmd>lua vim.lsp.buf.declaration()<cr>", { desc = "GoTo Declaration" })
	--	vim.keymap.set("n", "gI", "<cmd>lua vim.lsp.buf.implementation()<cr>", { desc = "GoTo Implementation" })
	--	vim.keymap.set("n", "gr", "<cmd>Telescope lsp_references <cr>", { desc = "GoTo References" })

	-- LSP Utils
	vim.keymap.set("n", "<space>la", "<cmd>lua vim.lsp.buf.code_action()<cr>", { desc = "Code Action" })
	vim.keymap.set("n", "<space>lr", "<cmd>lua vim.lsp.buf.rename()<cr>", { desc = "Rename Symbol" })

	-- Trigger Diagnostic Popup
	vim.keymap.set("n", "gl", get_diag, { desc = "Trigger Diagnostic Popup" })
end

return M
