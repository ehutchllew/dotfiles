local M = {}

-- Two important keymaps to use while in Telescope are:
-- - Insert mode: <c-/>
-- - Normal mode: ?
--
-- This opens a window that show you all keymaps for the
-- current telescope picker. Useful to discover what 
-- Telescope can do as well as how to do it.

local function config()
	return {
		extensions = {
			["ui-select"] = {
				require("telescope.themes").get_dropdown()
			}
		}
	}
end

M.setup = function()
	local ts = require("telescope")
	local c = config()
	
	ts.setup(c)

	-- Enable extensions if installed
	pcall(ts.load_extension, "fzf")
	pcall(ts.load_extension, "ui-select")
end


return M

