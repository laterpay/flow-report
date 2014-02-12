package main

import (
	"testing"
)

func TestParseBranches(t *testing.T) {

	content := `
* develop                                                ed3cb48 Added "styled" class to checkboxes and radios
  feature/merchant-transaction-tag                       35ff36a [behind 6] moving tests around to please angry god "jenkins"
  master                                                 02cd1ac [behind 26] Translationz! (cherry picked from commit 9afeb4f4c0b77982b9be4d43e4544b7a02871dff)
  remotes/origin/HEAD                                    -> origin/develop
  remotes/origin/develop                                 ed3cb48 Added "styled" class to checkboxes and radios
  remotes/origin/feature/checkbox_and_radio_styling      3f828e3 laterpay/laterpay#862 Added ezmark form styling to accept terms and conditions checkbox as an example
  remotes/origin/feature/demo-remove-2nd-purchase-dialog ad50d04 fix template override by referring to block.super, which I only just learned about
  remotes/origin/feature/django16                        bdbf458 adding test settings and updating to django 1.6
  remotes/origin/feature/flashes_js_refactoring          c4b4c43 Hope that fixes the problems in flash_purchased_before
  remotes/origin/feature/floppyforms_dialogs             2a369ea added floppyforms. added template for add_to_ppu_invoice_form dialog
  remotes/origin/feature/jm-fix-webcore-blocks           0986005 bulk update: bix blcok -> block
  remotes/origin/feature/merchant-transaction-tag        e121146 empty commit to trigger PR test build stuff on jenkins alrighy pls wrk very sweat
  remotes/origin/feature/shared_device_info_dialog       6280bf4 added shared device marking confirmation dialog
  remotes/origin/feature/skip-addtoinvoice               17ff787 refactoring a bit to make more sense (skjartansson)
  remotes/origin/feature/spiderman-integration           3de14ab Updates due to spiderman web API changes.
  remotes/origin/master                                  ba2d799 Merge branch 'release/2.6'
  remotes/origin/release/2.6                             74e609a [auto] bump setup.py version for release 2.6
    `

	ris := parseBranches(content)
	t.Logf("%+v\n", ris)

}
