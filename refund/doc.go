// Copyright 2016 qujianping. All rights reserved.
/*
 demo: 对于这样的订单数据
 order 308659296037157561730809
 promotions: 144388, 1468, 944
 0-sku: 141657751836299537700323 total: 4800
 1-sku: 234521120636300916515917 total: 4800
 2-sku: 242427288234593159795777 total: 25800
 3-sku: 279202322835359006281839 total: 19800
 4-sku: 285569205035359557138572 total: 15800
 5-sku: 326376247635360586382175 total: 18800
 6-sku: 365255984436301439767383 total: 4800
 7-sku: 417516447535520513091412 total: 13800
 8-sku: 59590949235522694731517 total: 28800
 9-sku: 6058281535888776171863 total: 4800
 10-sku: 83244962136318864318903 total: 4800
 其中,sku-3不承担优惠2,sku-10对于优惠1最大能承受的金额为12
 经过2次列置换1次矩形置换,输出:
		+--------------+---------------+-------------+------------+
		|              | PAY-0(144388) | PAY-1(1468) | PAY-2(944) |
		+--------------+---------------+-------------+------------+
		| SKU-0(4800)  |          4685 |          84 |         31 |
		| SKU-1(4800)  |          4721 |          48 |         31 |
		| SKU-2(25800) |         25376 |         258 |        166 |
		| SKU-3(19800) |         19601 |         199 |          0 |
		| SKU-4(15800) |         15540 |         158 |        102 |
		| SKU-5(18800) |         18491 |         188 |        121 |
		| SKU-6(4800)  |          4721 |          48 |         31 |
		| SKU-7(13800) |         13573 |         138 |         89 |
		| SKU-8(28800) |         28202 |         287 |        311 |
		| SKU-9(4800)  |          4721 |          48 |         31 |
		| SKU-10(4800) |          4757 |          12 |         31 |
		+--------------+---------------+-------------+------------+

 Explode优惠爆破分为3个步骤:
 1. 均分优惠:逐行将sku总价按优惠承担比例分配到每个优惠项,经过此步骤后,该表格所有行等式成立,即对每个SKU来说所有优惠项总和等于该SKU总价
 2. 列置换: 列头实际该优惠总额 - 表格按列求和为该优惠当前总额 = 优惠差值,找出优惠差值最大的列(正列)和优惠差值最小的列(为负数,负列),这两列称之为为互补列;从正列里选取"金额最大且可调"的sku和负列进行调整,正列SKU-val,负列SKU+val;循环调整直到所有列达到平衡或者找不到互补列报错退出;注意列置换会考虑不承担优惠的配置,并为其分配优惠金额0;由示意图可以看出: 列置换不影响行等式的平衡
 3. 矩形置换: 寻找超出配置最大值的某个SKU优惠分配点X,基于该点寻找调整矩形的A,B,C点构成调整矩形进行调整; 遍历表格单元格,先找到可调整的矩形对角点B点(大于0),然后检查矩形另外两个点是否满足调整调整(不超过配置最大值或无最大值配置),找到后进行一次矩形置换;循环进行矩形置换直到无超出配置值的点或找不到调整矩形报错退出;由示意图可以看出: 矩形置换对于行列等式的平衡均不影响

列置换示意图:
| ---- | ---------| ---- |
|  a1  |   .....  |  b1  |
|  a2  |   .....  |  b2  |
|  a3  |   .....  |  b3  |
|  A4  |   ==3=>  |  B4  |      A4为正列最大可调SKU,可调值为3(A4-3>=0), A4-=3, B4+=3
|  a5  |   .....  |  b5  |
|  a6  |   .....  |  b6  |      a列为正列,b列为负列

矩形置换示意图:
// A(+2)  <--------------- B(-2)
//   |                       |
//   |                       |
// X(-2)  ---------------> C(+2)

数学论证:
优惠分配问题实际上是一个m*n元一次方程组的求解: 例如对于m(m>=1)个SKU进行n(n>=1)项优惠分配,就是求解这样一组方程组:
X(i,0) + X(i,1) + ...... + X(i,n) = S(i) [S(i)为SKU(i)的价格,X(i,?)为S(i)分配到的具体优惠项的金额]
X(0,j) + X(1,j) + ...... + X(m,j) = P(j) [P(j)为优惠j的总优惠金额,X(?,j)为P(j)分配到具体SKU的优惠金额]
其中变量取值:  0<=i<=m, 0<=j<=n, S(i),P(j)>0 且满足 S(0) + S(1) + ...... + S(m) = P(0) + P(1) + ...... +P(n)
变量个数: m*n = m;
方程个数: m+n-1;这里看起来每行每列均有一个等式加上总的S=P,共 m+n+1 个等式,实则不然;最后的S=P可以由前面的等式推导出来(数据上天然构成造成的),而每行最后一个未知数又可以由前面的未知数表示成1-x的形式,所以最终会导致列方程减少1个,故而总的方程个数为: m+n+1-1-1 = m+n-1个
所以, 如果这个方程组有解,那么必须满足线性方程组有有限解(无解算做解集为0的有限解特例)的条件:

独立方程个数 >= 未知数个数
即, m+n-1>=mn  => m=1或n=1

基础结论: 在无附加条件的情况下,仅在一维向量的情况下有有限解;对应到业务上,即单SKU购买或多SKU单优惠项情况下有有限解,根据业务场景,这中情况不需要进行优惠拆分,天然成立.

考虑略复杂的情况,有时我们会增加额外的限定条件:某些SKU不允许承担优惠,或某些sku的某些优惠承担的最大值有限制;那么这些限制会为分解增加额外的f个限定方程,那么我们的有限解条件变成:

f+m+n-1>=mn

在该不等式限定的情况下,我们来讨论列置换算法的可行性:
* 如果在列置换中,我们找不到置换列对,说明方程组完全成立=>OK
* 如果在列置换中,我们找到一个正列,但是却找不到一个负列,那么此时的方程组可以退化为四元一次方程组(极简情形下为一维向量方程),m=n=2
| P0 | P1 |    |
| -- | -- | -- |
|  x | a  | S0 |
|  b | c  | S1 |
 x + a = S0
 b + c = S1
 x + b > P0
 c + a = P1
 左右相加很容易得出2x + 2a + 2b + 2c > 2x + 2a + 2b + 2c的悖论
所以,列置换的结果有两个:
1. 成功完成,使得方程组完全成立,优惠分配可行
2. 找不到完整互补列,无法分配

而导致结果2的原因就是:列置换的时候引入的某个分配X(i,j)=0的情况,每引入一个等式即增加了一个方程使得f-1,最终过度限制导致 f+m+n-1>=mn 无整数解.

此时我们再考虑其他限制存在情况下的矩形置换,即某些分配不允许超出某个最大优惠(max>0),对应业务场景是某个SKU分配的卡币不允许超过价格40%.
在经过列置换后,方程组已经保证了行列各自的完全平衡,如果发现某个分配点X超出了其配置最大限制,而又找不到一个置换矩形,反映到数学上就是无法找到一组整数解同时满足方程组和X点的不等式.该问题还是退化到四元方程组的简化问题上.example:
在列置换完成后的分解是这样的:
| 10 | 40 |    |
| -- | -- | -- |
| 10 | 10 | 20 |
|  0 | 30 | 30 |  这里SKU-1根据配置不允许承担优惠0
根据f+m+n-1>=mn, m=n=2 ===> f>=1,即如果再有一个有效限制条件就会使得方程组有唯一解或无解. 那么如果这个给出SKU-0能承担的优惠0最大为8,显然就导致了无解(找不到置换矩形)
所以,如果需要进行矩形置换,那么矩形置换的结果也有两个:
1. 成功完成,使得方程组完全成立(列置换已经保证,矩形置换不影响该平衡),优惠分配同时满足阈值限定
2. 找不到置换矩形,无法分配


在业务场景上,不论是列置换无法完成或者是矩形置换无法完成,都说明是第一步计算优惠出现问题,导致这里的分钱无解的情况。而列置换和矩形置换都能保证在有解的情况下求出唯一解或有效解.
*/
package refund
