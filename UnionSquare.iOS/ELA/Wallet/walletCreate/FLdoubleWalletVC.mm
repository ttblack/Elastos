//
//  FLdoubleWalletVC.m
//  FLWALLET
//
//  Created by fxl on 2018/4/20.
//  Copyright © 2018年 fxl. All rights reserved.
//

#import "FLdoubleWalletVC.h"
#import "FLPastWordVC.h"
#import "ELWalletManager.h"
#import "DAConfig.h"
@interface FLdoubleWalletVC ()
@property (weak, nonatomic) IBOutlet UIButton *PastWalletBtn;

/*
 *<# #>
 */
@property(copy,nonatomic)NSString *theMnemonicWordString;
@property (weak, nonatomic) IBOutlet UILabel *showInfoTextLabel;
@property (weak, nonatomic) IBOutlet UILabel *backupThePurseShowInfoTextLabel;
@end

@implementation FLdoubleWalletVC

- (void)viewDidLoad {
    [super viewDidLoad];
    [self defultWhite];
    self.title = NSLocalizedString(@"备份钱包", nil);
    self.Wallet.walletID=[NSString stringWithFormat:@"%@%@",@"wallet",[[FLTools share] getNowTimeTimestamp]];
    
    NSString *languageStringMword;

    NSString *languageString=[DAConfig userLanguage];
    if ([languageString  containsString:@"en"]) {
      languageStringMword=@"english";  
    }else if ([languageString  containsString:@"zh"]){
        
         languageStringMword=@"chinese";
    }
   
    
    invokedUrlCommand *mommand=[[invokedUrlCommand alloc]initWithArguments:@[languageStringMword] callbackId:self.Wallet.walletID className:@"Wallet" methodName:@"generateMnemonic"];
 

    
    [self.PastWalletBtn setTitle:NSLocalizedString(@"备份钱包", nil) forState:UIControlStateNormal];
    
    [self.PastWalletBtn setBackgroundColor:RGBA(255, 255, 255, 0.15) boldColor:[UIColor whiteColor] corner:0];
    
 PluginResult *result  = [[ELWalletManager share]generateMnemonic:mommand];
   self.Wallet.mnemonic=result.message[@"success"];
    self.showInfoTextLabel.text=NSLocalizedString(@"备份钱包:请将「助记词」抄写到安全的地方,干万不要保存在网络上。建议从转入,转出小额资产开始使用。", nil);
    self.backupThePurseShowInfoTextLabel.text=NSLocalizedString(@"立即备份您的钱包", nil);
    
}
- (IBAction)GoToPastWords:(id)sender {
    FLPastWordVC *vc = [[FLPastWordVC alloc]init];
    vc.Wallet=self.Wallet;
    
    [self.navigationController pushViewController:vc animated:YES];
}
-(void)viewWillAppear:(BOOL)animated
{
    [super viewWillAppear:animated];
    
}
@end
