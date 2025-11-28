# -*- coding: utf-8 -*-
"""
Created on Wed Apr 15 11:40:57 2020

@author: jamesyzhang
@计算auc
"""

from sklearn import metrics
import sys

if __name__ == "__main__":
    res_file = sys.argv[1]

    label = list()
    logits = list()
    with open(res_file, "r") as f:
        for line in f:
            tmp = line[0:-1].split(' ')
            label.append(int(tmp[0]))
            logits.append(float(tmp[1]))

    train_auc = metrics.roc_auc_score(label, logits)
    print(train_auc)
